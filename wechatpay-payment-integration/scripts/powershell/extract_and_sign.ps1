#
# 微信支付 APIv3 - P12 证书信息提取与签名工具 (Windows PowerShell 版)
#
# 推荐用法（签名时刻生成 TIMESTAMP / NONCE_STR）:
#   powershell -ExecutionPolicy Bypass -File extract_and_sign.ps1 `
#     -FilePath "apiclient_cert.p12" -Method GET -Url "/v3/pay/transactions/id/xxx?mchid=yyy"
#

param(
    [Parameter(Mandatory=$true, HelpMessage="Path to apiclient_cert.p12")]
    [string]$FilePath,

    [Parameter(Mandatory=$false, HelpMessage="P12 password (optional)")]
    [string]$Password = "",

    [Parameter(Mandatory=$false, HelpMessage="HTTP method, e.g. GET")]
    [string]$Method = "",

    [Parameter(Mandatory=$false, HelpMessage="Request URL path+query, starts with /")]
    [string]$Url = "",

    [Parameter(Mandatory=$false, HelpMessage="Unix timestamp seconds (optional)")]
    [string]$Timestamp = "",

    [Parameter(Mandatory=$false, HelpMessage="Nonce string (optional)")]
    [string]$NonceStr = "",

    [Parameter(Mandatory=$false, HelpMessage="Full sign string; use \\n for newlines")]
    [string]$SignString = ""
)

$ErrorActionPreference = "Stop"

function Get-MchidFromHint([string]$hint) {
    if ($hint -match '[\?&]mchid=([^&\\n]+)') { return $Matches[1] }
    if ($hint -match '[\?&]sp_mchid=([^&\\n]+)') { return $Matches[1] }
    return $null
}

if ([string]::IsNullOrEmpty($SignString)) {
    if ([string]::IsNullOrEmpty($Method) -or [string]::IsNullOrEmpty($Url)) {
        Write-Host "错误: 请提供 -Method 与 -Url，或提供 -SignString" -ForegroundColor Red
        exit 1
    }
    if (-not $Url.StartsWith("/")) {
        Write-Host "错误: -Url 必须以 / 开头（path + query）" -ForegroundColor Red
        exit 1
    }
    if ([string]::IsNullOrEmpty($Timestamp)) {
        $Timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds().ToString()
    }
    if ([string]::IsNullOrEmpty($NonceStr)) {
        $bytes = New-Object byte[] 16
        [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
        $NonceStr = ([BitConverter]::ToString($bytes) -replace '-', '')
    }
    $SignString = "$Method`n$Url`n$Timestamp`n$NonceStr`n`n"
} elseif ([string]::IsNullOrEmpty($Timestamp) -or [string]::IsNullOrEmpty($NonceStr)) {
    $lines = ($SignString -replace '\\n', "`n") -split "`n"
    if ($lines.Count -ge 4) {
        if ([string]::IsNullOrEmpty($Timestamp)) { $Timestamp = $lines[2] }
        if ([string]::IsNullOrEmpty($NonceStr)) { $NonceStr = $lines[3] }
    }
}

if ($FilePath.StartsWith("file://")) {
    $FilePath = $FilePath.Substring(7)
    if (-not $FilePath.StartsWith("/")) { $FilePath = "/" + $FilePath }
} elseif ($FilePath.StartsWith("file:")) {
    $FilePath = $FilePath.Substring(5)
}
if ($FilePath.StartsWith("@")) { $FilePath = $FilePath.Substring(1) }

if ($FilePath -like "/path/to/*" -or $FilePath -like "*\path\to\*" -or $FilePath -like "*:\path\to\*") {
    Write-Host "错误: 请将 -FilePath 参数替换为你本地 P12 证书的真实路径" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $FilePath)) {
    Write-Host "错误: P12 文件不存在: $FilePath" -ForegroundColor Red
    exit 1
}

function Test-PemLikeFile([string]$path) {
    $ext = [System.IO.Path]::GetExtension($path).ToLowerInvariant()
    if ($ext -in '.pem', '.crt', '.cer', '.key') { return $true }
    $first = Get-Content -Path $path -TotalCount 1 -ErrorAction SilentlyContinue
    return ($first -like '-----BEGIN*')
}

function Write-WrongCertFormatError([string]$path) {
    $name = [System.IO.Path]::GetFileName($path)
    Write-Host "错误: -FilePath 指向的是 PEM/证书文件（$name），本脚本仅支持 PKCS#12 格式的 apiclient_cert.p12" -ForegroundColor Red
    Write-Host "提示: 请将 -FilePath 改为你本地的 apiclient_cert.p12 路径（证书压缩包内通常同时提供 .p12 与 .pem，请选用 .p12）" -ForegroundColor Yellow
    exit 1
}

if (Test-PemLikeFile $FilePath) {
    Write-WrongCertFormatError $FilePath
}

$P12FullPath = (Resolve-Path $FilePath).Path
$pwdHint = "$Url$SignString"

try {
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2(
        $P12FullPath, $Password,
        [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::Exportable
    )
} catch {
    if ([string]::IsNullOrEmpty($Password)) {
        $cand = Get-MchidFromHint $pwdHint
        if ($cand) {
            try {
                $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2(
                    $P12FullPath, $cand,
                    [System.Security.Cryptography.X509Certificates.X509KeyStorageFlags]::Exportable
                )
                $Password = $cand
            } catch {
                Write-Host '错误: 无法加载 P12 文件。可能需要 P12 密码，请在命令中追加: -Password "你的P12密码"（常见为商户号）' -ForegroundColor Red
                exit 1
            }
        } else {
            Write-Host '错误: 无法加载 P12 文件。可能需要 P12 密码，请在命令中追加: -Password "你的P12密码"（常见为商户号）' -ForegroundColor Red
            exit 1
        }
    } else {
        if (Test-PemLikeFile $FilePath) {
            Write-WrongCertFormatError $FilePath
        }
        Write-Host "错误: 无法加载 P12 文件，请检查 -Password 是否为 P12 密码（常见为商户号）" -ForegroundColor Red
        exit 1
    }
}

$serial = $cert.SerialNumber
if ([string]::IsNullOrEmpty($serial)) {
    Write-Host "错误: 无法提取证书序列号" -ForegroundColor Red
    exit 1
}

$subject = $cert.Subject
if ($subject -match 'CN=([^,]+)') { $mchid = $Matches[1].Trim() }
else { $mchid = "（无法从证书 CN 字段提取，请手动确认）" }

$signContent = $SignString -replace '\\n', "`n"
$bytes = [System.Text.Encoding]::UTF8.GetBytes($signContent)
$signatureBytes = $null

# Windows PowerShell 5.1 / .NET Framework：使用 PrivateKey + SignHash
$legacyRsa = $cert.PrivateKey
if ($null -ne $legacyRsa) {
    try {
        $sha256 = New-Object System.Security.Cryptography.SHA256CryptoServiceProvider
        $hash = $sha256.ComputeHash($bytes)
        $oid = [System.Security.Cryptography.CryptoConfig]::MapNameToOID("SHA256")
        $signatureBytes = $legacyRsa.SignHash($hash, $oid)
    } catch {
        $signatureBytes = $null
    } finally {
        if ($null -ne $sha256) { $sha256.Dispose() }
    }
}

# PowerShell 7+ / .NET 4.6+：使用 RSACertificateExtensions::GetRSAPrivateKey
if ($null -eq $signatureBytes) {
    try {
        $rsaType = [System.Security.Cryptography.X509Certificates.RSACertificateExtensions]
        $rsa = $rsaType::GetRSAPrivateKey($cert)
        if ($null -ne $rsa) {
            $signatureBytes = $rsa.SignData($bytes,
                [System.Security.Cryptography.HashAlgorithmName]::SHA256,
                [System.Security.Cryptography.RSASignaturePadding]::Pkcs1)
        }
    } catch {
        $signatureBytes = $null
    }
}

if ($null -eq $signatureBytes) {
    Write-Host "错误: 无法提取私钥或签名失败，请确认 P12 文件包含私钥" -ForegroundColor Red
    exit 1
}

try {
    $signature = [Convert]::ToBase64String($signatureBytes)
} catch {
    Write-Host "错误: 签名失败" -ForegroundColor Red
    exit 1
}

Write-Host "---------- 签名结果开始 ----------" -ForegroundColor Green
Write-Host "API 证书序列号（serial_no）: $serial"
Write-Host "时间戳（timestamp）: $Timestamp"
Write-Host "随机串（nonce_str）: $NonceStr"
Write-Host "API 证书中的商户号: $mchid"
Write-Host "签名值（signature）: $signature"
Write-Host "---------- 签名结果结束 ----------" -ForegroundColor Green
