"""
wechatpay-docs-sync.py — 微信支付知识库远程同步工具

用法:
    python3 wechatpay-docs-sync.py update  # 检查远程并同步（含首次安装）
    python3 wechatpay-docs-sync.py --encoding utf-8 update  # 显式锁定 UTF-8（推荐 Windows）
"""

from __future__ import annotations

import json
import locale
import os
import shutil
import stat
import sys
import tarfile
import tempfile
import zipfile
from datetime import datetime, timedelta, timezone
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

VERSION = "1.0"

# ==== 脚本配置 ====

SCRIPT_DIR = Path(__file__).resolve().parent  # scripts directory/
SKILL_DIR = SCRIPT_DIR.parent  # skill directory/

DOCS_URL = "https://wx.gtimg.com/resource/wechatpay_api/wechatpay-docs.zip"
DOCS_TARGET_DIR = SKILL_DIR / "assets"
STATE_FILE = SCRIPT_DIR / ".wechatpay-docs-sync-state.json"
CHECK_INTERVAL_HOURS = 12

_TZ = timezone(timedelta(hours=8))

# ==== HTTP header（统一小写） ====

# 代理/网关可能会改写 Header 字段名大小写；字段名本身大小写不敏感。
# 这里统一将“字段名”转为大写进行匹配，避免因大小写变化导致取值失败。
H_LAST_MODIFIED = "LAST-MODIFIED"
H_ETAG = "ETAG"
H_CONTENT_LENGTH = "CONTENT-LENGTH"

# ==== 状态文件 key ====

S_LAST_CHECK = "LAST_CHECK_TIME"
S_LAST_UPDATE = "LAST_UPDATE_TIME"
S_REMOTE_MODIFIED = "REMOTE_LAST_MODIFIED"
S_REMOTE_ETAG = "REMOTE_ETAG"


USER_AGENT = f"{Path(__file__).stem}/{VERSION}"
IGNORED_FILES = {".DS_Store", "Thumbs.db", "__MACOSX"}
DOCS_FILE_GLOB = "*.md"
ZIP_FALLBACK_ENCODING = "cp437"
DEFAULT_IO_ENCODING = "utf-8"


# ==== 编码 ====


def _set_windows_console_utf8() -> None:
    if sys.platform != "win32":
        return
    try:
        import ctypes

        kernel32 = ctypes.windll.kernel32  # type: ignore[attr-defined]
        kernel32.SetConsoleOutputCP(65001)
        kernel32.SetConsoleCP(65001)
    except (OSError, AttributeError):
        pass


def _bootstrap_encoding(encoding: str = DEFAULT_IO_ENCODING) -> None:
    """锁定脚本 I/O 编码，降低系统区域/终端代码页对中文路径与输出的干扰。"""
    normalized = (encoding or DEFAULT_IO_ENCODING).strip().lower()
    os.environ["PYTHONIOENCODING"] = f"{normalized}:replace"

    for stream in (sys.stdout, sys.stderr):
        if hasattr(stream, "reconfigure"):
            try:
                stream.reconfigure(encoding=normalized, errors="replace")
            except (OSError, ValueError, AttributeError):
                pass

    for name in ("en_US.UTF-8", "C.UTF-8", "UTF-8"):
        try:
            locale.setlocale(locale.LC_ALL, name)
            break
        except locale.Error:
            continue

    if sys.platform == "win32":
        _set_windows_console_utf8()


def _parse_cli_args(argv: list[str]) -> tuple[str, str]:
    """解析可选 --encoding / -E，返回 (command, encoding)。"""
    encoding = DEFAULT_IO_ENCODING
    args = list(argv)
    i = 0
    while i < len(args):
        token = args[i]
        if token in ("--encoding", "-E") and i + 1 < len(args):
            encoding = args[i + 1]
            del args[i : i + 2]
            continue
        if token.startswith("--encoding="):
            encoding = token.split("=", 1)[1]
            del args[i]
            continue
        i += 1
    return (args[0] if args else "", encoding)


# ==== 状态管理 ====


def _now_iso() -> str:
    """返回当前时间的 ISO 8601 字符串（东八区）。"""
    return datetime.now(_TZ).isoformat(timespec="seconds")


def _load_state() -> dict:
    """从 STATE_FILE 读取同步状态，文件不存在则返回空 dict。"""
    if STATE_FILE.exists():
        state = json.loads(STATE_FILE.read_text(encoding="utf-8"))
        # 兼容：部分环境可能会改写 header value 的大小写；本地 state 统一按小写存取。
        if isinstance(state.get(S_REMOTE_ETAG), str):
            state[S_REMOTE_ETAG] = state[S_REMOTE_ETAG].lower()
        if isinstance(state.get(S_REMOTE_MODIFIED), str):
            state[S_REMOTE_MODIFIED] = state[S_REMOTE_MODIFIED].lower()
        return state
    return {}


def _save_state(state: dict) -> None:
    """将同步状态写入 STATE_FILE。"""
    STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
    STATE_FILE.write_text(
        json.dumps(state, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )


# ==== 远程资源 ====


def _head_remote() -> dict:
    """HEAD 请求，返回 last_modified / etag / content_length 或 error。"""
    req = Request(DOCS_URL, method="HEAD")
    req.add_header("User-Agent", USER_AGENT)
    try:
        with urlopen(req, timeout=15) as resp:
            headers = {k.upper(): v for k, v in resp.headers.items()}
            # value 统一小写用于比较/落盘，避免被代理改写大小写导致误判
            lm = headers.get(H_LAST_MODIFIED)
            etag = headers.get(H_ETAG)
            return {
                H_LAST_MODIFIED: lm.lower() if isinstance(lm, str) else lm,
                H_ETAG: etag.lower() if isinstance(etag, str) else etag,
                H_CONTENT_LENGTH: headers.get(H_CONTENT_LENGTH),
            }
    except (URLError, HTTPError) as exc:
        return {"error": str(exc)}


def _within_interval(state: dict) -> bool:
    """判断距上次检查是否不足 CHECK_INTERVAL_HOURS 小时。"""
    ts = state.get(S_LAST_CHECK)
    if not ts:
        return False
    elapsed = datetime.now(_TZ) - datetime.fromisoformat(ts)
    return elapsed.total_seconds() < CHECK_INTERVAL_HOURS * 3600


def _remote_changed(state: dict, remote: dict) -> bool:
    """比较远程 ETag / Last-Modified 与本地记录，判断是否有变化。"""
    r_etag = remote.get(H_ETAG)
    if r_etag and state.get(S_REMOTE_ETAG):
        return str(r_etag).lower() != str(state[S_REMOTE_ETAG]).lower()
    r_lm = remote.get(H_LAST_MODIFIED)
    if r_lm and state.get(S_REMOTE_MODIFIED):
        return str(r_lm).lower() != str(state[S_REMOTE_MODIFIED]).lower()
    return True  # 无法判断时视为有变化


# ==== 下载与解压 ====


def _download(dest: Path) -> None:
    """从 DOCS_URL 下载压缩包到 dest，显示下载进度。"""
    req = Request(DOCS_URL)
    req.add_header("User-Agent", USER_AGENT)
    with urlopen(req, timeout=300) as resp:
        headers = {k.upper(): v for k, v in resp.headers.items()}
        total = headers.get(H_CONTENT_LENGTH)
        total = int(total) if total else None
        done = 0
        last_pct = -1
        last_mb = -1
        is_tty = sys.stdout.isatty()
        with open(dest, "wb") as fp:
            while True:
                chunk = resp.read(65536)
                if not chunk:
                    break
                fp.write(chunk)
                done += len(chunk)
                if total:
                    pct = done * 100 // total
                    if pct == last_pct:
                        continue
                    last_pct = pct
                    # 非 TTY（IDE 输出区、管道等）下 \r 无法覆写同行，按里程碑换行
                    if not is_tty and pct % 10 != 0 and pct < 100:
                        continue
                    msg = (
                        f"  下载: {done / 1048576:.1f} MB / "
                        f"{total / 1048576:.1f} MB ({pct}%)"
                    )
                    if is_tty:
                        print(f"\r{msg}", end="", flush=True)
                    else:
                        print(msg, flush=True)
                else:
                    mb = int(done / 1048576)
                    if is_tty:
                        print(f"\r  下载: {done / 1048576:.1f} MB", end="", flush=True)
                    elif mb > last_mb:
                        last_mb = mb
                        print(f"  下载: {done / 1048576:.1f} MB", flush=True)
        print()


def _cjk_char_count(text: str) -> int:
    return sum(1 for ch in text if "\u4e00" <= ch <= "\u9fff")


def _try_recover_mojibake(name: str) -> str:
    """UTF-8 文件名被误按 Latin-1/CP1252 解码时，尝试 latin-1→utf-8 还原。"""
    try:
        recovered = name.encode("latin-1").decode("utf-8")
    except (UnicodeDecodeError, UnicodeEncodeError):
        return name
    if _cjk_char_count(recovered) > _cjk_char_count(name):
        return recovered
    return name


def _decode_zip_filename(info: zipfile.ZipInfo) -> str:
    """跨平台还原 zip 内文件名（Windows 上 zf.extract 偶发中文乱码，须先修正）。"""
    name = info.filename.replace("\\", "/")
    if info.flag_bits & 0x800:
        return _try_recover_mojibake(name)
    for encoding in ("utf-8", "gbk", "gb18030"):
        try:
            decoded = name.encode(ZIP_FALLBACK_ENCODING).decode(encoding)
            if _cjk_char_count(decoded) >= _cjk_char_count(name):
                return decoded
        except (UnicodeDecodeError, UnicodeEncodeError):
            continue
    return _try_recover_mojibake(name)


def _win_long_path(path: Path) -> str:
    """Windows 长路径前缀，绕过 MAX_PATH(260) 限制。"""
    resolved = str(path.resolve())
    if resolved.startswith("\\\\?\\"):
        return resolved
    if resolved.startswith("\\\\"):
        return "\\\\?\\UNC\\" + resolved[2:]
    return "\\\\?\\" + resolved


def _mkdir_parents(path: Path) -> None:
    try:
        path.mkdir(parents=True, exist_ok=True)
    except OSError:
        if sys.platform != "win32":
            raise
        os.makedirs(_win_long_path(path), exist_ok=True)


def _write_bytes(path: Path, data: bytes) -> None:
    _mkdir_parents(path.parent)
    try:
        path.write_bytes(data)
    except OSError:
        if sys.platform != "win32":
            raise
        with open(_win_long_path(path), "wb") as fp:
            fp.write(data)


def _extract_zip(archive: Path, dest: Path) -> None:
    """手动解压 zip，避免 Windows 上 ZipFile.extract 写出乱码中文路径。"""
    dest.mkdir(parents=True, exist_ok=True)
    zip_kwargs: dict = {}
    if sys.version_info >= (3, 11):
        zip_kwargs["metadata_encoding"] = "utf-8"
    with zipfile.ZipFile(archive, **zip_kwargs) as zf:
        for info in zf.infolist():
            name = _decode_zip_filename(info)
            if not name or name.startswith("__MACOSX"):
                continue
            parts = name.split("/")
            if any(p in IGNORED_FILES for p in parts):
                continue
            target = dest.joinpath(*parts)
            if name.endswith("/") or info.is_dir():
                _mkdir_parents(target)
                continue
            with zf.open(info) as src:
                _write_bytes(target, src.read())


def _extract(archive: Path, dest: Path) -> None:
    """解压 zip 或 tar.gz 到 dest。"""
    dest.mkdir(parents=True, exist_ok=True)
    if zipfile.is_zipfile(archive):
        _extract_zip(archive, dest)
        return
    try:
        with tarfile.open(archive) as tf:
            tf.extractall(dest, filter="data")
        return
    except (tarfile.TarError, TypeError):
        pass
    raise RuntimeError(
        f"无法识别压缩格式，请确认下载链接是否为 zip 或 tar.gz 文件: {archive.name}"
    )


def _find_content_root(extract_dir: Path) -> Path:
    """若解压后只有单一顶层目录，则进入该目录作为实际内容根。"""
    items = [p for p in extract_dir.iterdir() if p.name != "__MACOSX"]
    if len(items) == 1 and items[0].is_dir():
        return items[0]
    return extract_dir


def _chmod_writable(path: Path) -> None:
    try:
        os.chmod(path, stat.S_IWRITE)
    except OSError:
        pass


def _on_rm_error(func, path, _exc_info) -> None:
    """Windows 上只读文件/目录删除失败时，先改权限再重试。"""
    _chmod_writable(Path(path))
    func(path)


def _unlink(path: Path) -> None:
    """删除单个文件或符号链接。"""
    _chmod_writable(path)
    try:
        path.unlink()
        return
    except OSError:
        if sys.platform != "win32":
            raise
    os.remove(_win_long_path(path))


def _rmdir(path: Path) -> None:
    """删除空目录。"""
    _chmod_writable(path)
    try:
        path.rmdir()
        return
    except OSError:
        if sys.platform != "win32":
            raise
    os.rmdir(_win_long_path(path))


def _remove_tree(root: Path) -> None:
    """删除目录树；Windows 下自底向上并使用长路径，避免 rmtree 在深层目录失败。"""
    if not root.exists():
        return
    if sys.platform == "win32":
        walk_root = _win_long_path(root.resolve())
        for dirpath, dirnames, filenames in os.walk(walk_root, topdown=False):
            for name in filenames:
                fp = os.path.join(dirpath, name)
                try:
                    _chmod_writable(Path(fp))
                    os.remove(fp)
                except OSError:
                    try:
                        os.remove(_win_long_path(Path(fp)))
                    except OSError as exc:
                        raise OSError(f"无法删除文件: {fp}") from exc
            for name in dirnames:
                dp = os.path.join(dirpath, name)
                try:
                    _chmod_writable(Path(dp))
                    os.rmdir(dp)
                except OSError:
                    try:
                        os.rmdir(_win_long_path(Path(dp)))
                    except OSError as exc:
                        raise OSError(f"无法删除目录: {dp}") from exc
        try:
            _rmdir(root)
        except OSError as exc:
            raise OSError(f"无法删除目录: {root}") from exc
        return
    shutil.rmtree(root, onerror=_on_rm_error)


def _copy_file(src: Path, dst: Path) -> None:
    """复制单个文件；Windows 下自动尝试长路径。"""
    dst.parent.mkdir(parents=True, exist_ok=True)
    try:
        shutil.copy2(src, dst)
        return
    except OSError:
        if sys.platform != "win32":
            raise
    shutil.copy2(_win_long_path(src), _win_long_path(dst))


def _copy_tree(src: Path, dst: Path) -> list[tuple[Path, str]]:
    """逐文件复制目录树，返回 (相对路径, 错误信息) 列表。"""
    dst.mkdir(parents=True, exist_ok=True)
    errors: list[tuple[Path, str]] = []
    for item in sorted(src.rglob("*")):
        if item.name in IGNORED_FILES or item.name == "__MACOSX":
            continue
        rel = item.relative_to(src)
        target = dst / rel
        if item.is_dir():
            try:
                target.mkdir(parents=True, exist_ok=True)
            except OSError as exc:
                if sys.platform == "win32":
                    try:
                        os.mkdir(_win_long_path(target), exist_ok=True)
                    except OSError as exc2:
                        errors.append((rel, str(exc2)))
                else:
                    errors.append((rel, str(exc)))
            continue
        try:
            _copy_file(item, target)
        except OSError as exc:
            errors.append((rel, str(exc)))
    return errors


def _clear_docs_dir() -> None:
    """删除 DOCS_TARGET_DIR 下的所有文件与子目录（保留 assets 目录本身）。"""
    if not DOCS_TARGET_DIR.exists():
        return
    resolved = DOCS_TARGET_DIR.resolve()
    skill_root = SKILL_DIR.resolve()
    if not str(resolved).startswith(str(skill_root)):
        raise RuntimeError(f"目标目录不在 skill 范围内，拒绝清空: {DOCS_TARGET_DIR}")
    for child in DOCS_TARGET_DIR.iterdir():
        if child.is_symlink():
            _unlink(child)
        elif child.is_dir():
            _remove_tree(child)
        else:
            _unlink(child)


# ==== 命令 ====


def _is_installed() -> bool:
    """判断本地知识库目录是否存在且非空。"""
    return DOCS_TARGET_DIR.exists() and any(DOCS_TARGET_DIR.iterdir())


def cmd_update() -> None:
    """检查远程版本；有更新或首次安装时下载、清空 assets 并写入新版本。"""
    state = _load_state()
    installed = _is_installed()

    if installed and _within_interval(state):
        last = state.get(S_LAST_CHECK, "未知")
        print(
            f"知识库已是最新（上次检查: {last}，{CHECK_INTERVAL_HOURS}H 内无需重复检查）。"
        )
        return

    print("正在检查远程文档版本…")
    remote = _head_remote()
    if "error" in remote:
        print(
            f"无法连接远程服务器，请检查网络后重试。\n  错误详情: {remote['error']}",
            file=sys.stderr,
        )
        sys.exit(1)

    state[S_LAST_CHECK] = _now_iso()

    if installed and not _remote_changed(state, remote):
        _save_state(state)
        print("远程文档未发生变化，当前已是最新，无需更新。")
        return

    if not installed:
        print("本地尚未安装知识库，开始首次下载…")
    else:
        print(
            f"检测到远程文档已更新（{remote.get(H_LAST_MODIFIED, '时间未知')}），开始下载…"
        )

    with tempfile.TemporaryDirectory() as tmp:
        tmp = Path(tmp)
        suffix = Path(DOCS_URL.split("?")[0]).suffix or ".zip"
        archive = tmp / f"docs{suffix}"

        _download(archive)

        print("下载完成，正在解压…")
        extract_dir = tmp / "out"
        _extract(archive, extract_dir)
        new_root = _find_content_root(extract_dir)

        print("正在写入知识库...")
        DOCS_TARGET_DIR.parent.mkdir(parents=True, exist_ok=True)
        _clear_docs_dir()
        copy_errors = _copy_tree(new_root, DOCS_TARGET_DIR)
        if copy_errors:
            print(
                f"写入知识库时出现 {len(copy_errors)} 个文件错误（常见于 Windows 长路径或权限问题）：",
                file=sys.stderr,
            )
            for rel, msg in copy_errors[:10]:
                print(f"  - {rel}: {msg}", file=sys.stderr)
            if len(copy_errors) > 10:
                print(f"  … 另有 {len(copy_errors) - 10} 个错误未列出", file=sys.stderr)
            sys.exit(1)

    state[S_LAST_UPDATE] = _now_iso()
    # 为了抵抗代理/网关对 header value 的大小写改写，这里落盘时统一做小写。
    state[S_REMOTE_MODIFIED] = str(remote.get(H_LAST_MODIFIED, "") or "").lower()
    state[S_REMOTE_ETAG] = str(remote.get(H_ETAG, "") or "").lower()
    _save_state(state)

    count = sum(1 for _ in DOCS_TARGET_DIR.rglob(DOCS_FILE_GLOB))
    print(f"更新完成，当前共 {count} 篇文档。")


# ==== 入口 ====

_USAGE = """\
用法: python3 wechatpay-docs-sync.py [--encoding utf-8] update

  update              检查远程是否有更新；有变化时下载并全量替换本地知识库（含首次安装）
                      默认 12 小时内不重复检查远程
  --encoding, -E ENC  锁定脚本 stdout/stderr 编码（默认 utf-8；Windows 推荐显式指定）

示例（Windows）:
  python wechatpay-docs-sync.py --encoding utf-8 update
  set PYTHONUTF8=1 && python wechatpay-docs-sync.py update"""


def main(argv: list[str] | None = None) -> None:
    raw = list(argv if argv is not None else sys.argv[1:])
    if not raw or raw[0] in ("-h", "--help"):
        print(_USAGE)
        sys.exit(0)

    cmd, encoding = _parse_cli_args(raw)
    _bootstrap_encoding(encoding)

    if cmd != "update":
        print(f"未识别到有效命令: {cmd or '(空)'}\n")
        print(_USAGE)
        sys.exit(1)

    cmd_update()


if __name__ == "__main__":
    main()
