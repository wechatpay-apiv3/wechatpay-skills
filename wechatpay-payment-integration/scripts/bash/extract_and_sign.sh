#!/bin/bash
#
# 微信支付 APIv3 - P12 证书信息提取与签名工具
#
# 用法（推荐，签名时刻生成 TIMESTAMP / NONCE_STR）:
#   bash extract_and_sign.sh --filePath <P12> [--password <密码>] \
#     --method GET --url '/v3/pay/transactions/id/xxx?mchid=yyy'
#
# 用法（可选，自行提供时间戳与随机串）:
#   bash extract_and_sign.sh --filePath <P12> --method GET --url '...' \
#     --timestamp <秒级时间戳> --nonce_str <32位随机串>
#
# 用法（兼容，传入完整待签名串）:
#   bash extract_and_sign.sh --filePath <P12> --signString 'GET\n/...\n...\n\n'

set -euo pipefail

P12_FILE=""
P12_PASSWORD=""
HTTP_METHOD=""
REQUEST_URL=""
TIMESTAMP_IN=""
NONCE_IN=""
SIGN_STRING=""

usage() {
    echo "用法: bash extract_and_sign.sh --filePath <P12文件路径> [--password <P12密码>] \\"
    echo "        (--method <HTTP方法> --url '<请求URL路径+query>') | --signString '<待签名串>'"
    echo ""
    echo "参数说明:"
    echo "  --filePath      apiclient_cert.p12 文件路径（支持 file:///... 或 @/path/...）"
    echo "  --password      P12 密码（可选；未传时可能从 URL 中的 mchid/sp_mchid 尝试）"
    echo "  --method        HTTP 方法，如 GET（与 --url 搭配使用）"
    echo "  --url           请求 URL（含 path 与 query，以 / 开头）"
    echo "  --timestamp     秒级时间戳（可选；未传则在签名时自动生成）"
    echo "  --nonce_str     随机串（可选；未传则在签名时自动生成）"
    echo "  --signString    完整待签名串（兼容旧用法；含 \\\\n 转义换行）"
    exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --filePath)
      P12_FILE="${2:-}"
      shift 2
      ;;
    --password)
      P12_PASSWORD="${2:-}"
      shift 2
      ;;
    --method)
      HTTP_METHOD="${2:-}"
      shift 2
      ;;
    --url)
      REQUEST_URL="${2:-}"
      shift 2
      ;;
    --timestamp)
      TIMESTAMP_IN="${2:-}"
      shift 2
      ;;
    --nonce_str)
      NONCE_IN="${2:-}"
      shift 2
      ;;
    --signString)
      SIGN_STRING="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "错误: 未知参数: $1"
      usage
      ;;
  esac
done

if [ -z "$P12_FILE" ]; then
    usage
fi

if [ -z "$SIGN_STRING" ]; then
    if [ -z "$HTTP_METHOD" ] || [ -z "$REQUEST_URL" ]; then
        echo "错误: 请提供 --method 与 --url，或提供 --signString"
        usage
    fi
    if [[ "$REQUEST_URL" != /* ]]; then
        echo "错误: --url 必须以 / 开头（path + query）"
        exit 1
    fi
    if [ -z "$TIMESTAMP_IN" ]; then
        TIMESTAMP_IN="$(date +%s)"
    fi
    if [ -z "$NONCE_IN" ]; then
        NONCE_IN="$(openssl rand -hex 16 | tr '[:lower:]' '[:upper:]')"
    fi
    SIGN_STRING="${HTTP_METHOD}"$'\n'"${REQUEST_URL}"$'\n'"${TIMESTAMP_IN}"$'\n'"${NONCE_IN}"$'\n\n'
elif [ -z "$TIMESTAMP_IN" ] || [ -z "$NONCE_IN" ]; then
  # 从完整待签名串解析时间戳与随机串（第 3、4 行）
  [ -z "$TIMESTAMP_IN" ] && TIMESTAMP_IN=$(printf '%b' "$SIGN_STRING" | sed -n '3p')
  [ -z "$NONCE_IN" ] && NONCE_IN=$(printf '%b' "$SIGN_STRING" | sed -n '4p')
fi

# 兼容 file:/// file: @ 前缀
if [[ "$P12_FILE" == file://* ]]; then
    P12_FILE="${P12_FILE#file://}"
    if [[ "$P12_FILE" != /* ]]; then
        P12_FILE="/$P12_FILE"
    fi
elif [[ "$P12_FILE" == file:* ]]; then
    P12_FILE="${P12_FILE#file:}"
fi
if [[ "$P12_FILE" == @* ]]; then
    P12_FILE="${P12_FILE#@}"
fi

if [[ "$P12_FILE" == "/path/to/"* ]] || [[ "$P12_FILE" == *"\\path\\to\\"* ]] || [[ "$P12_FILE" == *":\\path\\to\\"* ]]; then
    echo "错误: 请将 --filePath 参数替换为你本地 P12 证书的真实路径"
    exit 1
fi

if [ ! -f "$P12_FILE" ]; then
    echo "错误: P12 文件不存在: $P12_FILE"
    exit 1
fi

_is_pem_like_file() {
    local file="$1"
    local ext first
    ext="${file##*.}"
    ext=$(printf '%s' "$ext" | tr '[:upper:]' '[:lower:]')
    case "$ext" in
        pem|crt|cer|key) return 0 ;;
    esac
    first=$(head -n 1 "$file" 2>/dev/null || true)
    [[ "$first" == -----BEGIN* ]]
}

_fail_wrong_cert_format() {
    local file="$1"
    echo "错误: --filePath 指向的是 PEM/证书文件（$(basename "$file")），本脚本仅支持 PKCS#12 格式的 apiclient_cert.p12"
    echo "提示: 请将 --filePath 改为你本地的 apiclient_cert.p12 路径（证书压缩包内通常同时提供 .p12 与 .pem，请选用 .p12）"
    exit 1
}

if _is_pem_like_file "$P12_FILE"; then
    _fail_wrong_cert_format "$P12_FILE"
fi

_guess_p12_password() {
    local hint="$1"
    local pwd
    pwd=$(printf '%s' "$hint" | sed -n 's/.*[?&]mchid=\([^&]*\).*/\1/p' | head -n 1)
    if [ -z "$pwd" ]; then
        pwd=$(printf '%s' "$hint" | sed -n 's/.*[?&]sp_mchid=\([^&]*\).*/\1/p' | head -n 1)
    fi
    printf '%s' "$pwd"
}

# 密码候选：优先从 URL 提取 mchid / sp_mchid
PWD_HINT="${REQUEST_URL:-$SIGN_STRING}"

LEGACY_FLAG=""
PASSIN="pass:$P12_PASSWORD"

if openssl pkcs12 -in "$P12_FILE" -clcerts -nokeys -passin "$PASSIN" -legacy >/dev/null 2>&1; then
    LEGACY_FLAG="-legacy"
elif ! openssl pkcs12 -in "$P12_FILE" -clcerts -nokeys -passin "$PASSIN" >/dev/null 2>&1; then
    if [ -z "$P12_PASSWORD" ]; then
        CAND_PWD="$(_guess_p12_password "$PWD_HINT")"
        if [ -n "$CAND_PWD" ]; then
            P12_PASSWORD="$CAND_PWD"
            PASSIN="pass:$P12_PASSWORD"
            if openssl pkcs12 -in "$P12_FILE" -clcerts -nokeys -passin "$PASSIN" -legacy >/dev/null 2>&1; then
                LEGACY_FLAG="-legacy"
            elif ! openssl pkcs12 -in "$P12_FILE" -clcerts -nokeys -passin "$PASSIN" >/dev/null 2>&1; then
                echo "错误: 无法读取 P12 文件。可能需要 P12 密码，请在命令中追加: --password \"<P12密码>\"（常见为商户号）"
                exit 1
            fi
        else
            echo "错误: 无法读取 P12 文件。可能需要 P12 密码，请在命令中追加: --password \"<P12密码>\"（常见为商户号）"
            exit 1
        fi
    else
        if _is_pem_like_file "$P12_FILE"; then
            _fail_wrong_cert_format "$P12_FILE"
        fi
        echo "错误: 无法读取 P12 文件，请检查 --password 是否为 P12 密码（常见为商户号）"
        exit 1
    fi
fi

CERT_PEM=$(openssl pkcs12 -in "$P12_FILE" -clcerts -nokeys -passin "$PASSIN" $LEGACY_FLAG 2>/dev/null)

SERIAL=$(echo "$CERT_PEM" | openssl x509 -serial -noout 2>/dev/null | sed 's/serial=//')
if [ -z "$SERIAL" ]; then
    echo "错误: 无法提取证书序列号"
    exit 1
fi

SUBJECT=$(echo "$CERT_PEM" | openssl x509 -subject -noout -nameopt RFC2253 2>/dev/null)
MCHID=$(echo "$SUBJECT" | sed -n 's/.*CN=\([^,]*\).*/\1/p')
if [ -z "$MCHID" ]; then
    MCHID="（无法从证书 CN 字段提取，请手动确认）"
fi

PRIVKEY=$(openssl pkcs12 -in "$P12_FILE" -nocerts -nodes -passin "$PASSIN" $LEGACY_FLAG 2>/dev/null)
if [ -z "$PRIVKEY" ]; then
    echo "错误: 无法提取私钥"
    exit 1
fi

SIGNATURE=$(printf "%b" "$SIGN_STRING" | \
    openssl dgst -sha256 -sign <(echo "$PRIVKEY") 2>/dev/null | \
    openssl base64 -A)

if [ -z "$SIGNATURE" ]; then
    echo "错误: 签名失败"
    exit 1
fi

echo "---------- 签名结果开始 ----------"
echo "API 证书序列号（serial_no）: $SERIAL"
echo "时间戳（timestamp）: $TIMESTAMP_IN"
echo "随机串（nonce_str）: $NONCE_IN"
echo "API 证书中的商户号: $MCHID"
echo "签名值（signature）: $SIGNATURE"
echo "---------- 签名结果结束 ----------"
