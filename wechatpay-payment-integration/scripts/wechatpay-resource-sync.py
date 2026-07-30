"""
wechatpay-resource-sync.py — 微信支付 Skill 及知识库远程同步工具

用法:
    python3 wechatpay-resource-sync.py update  # 检查远程并同步 Skill 和知识库（含首次安装）
"""

import json
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

VERSION = "1.2"

# ==== 脚本配置 ====

SCRIPT_DIR = Path(__file__).resolve().parent  # scripts directory/
SKILL_DIR = SCRIPT_DIR.parent  # skill directory/

DOCS_URL = "https://wx.gtimg.com/resource/wechatpay_api/wechatpay-docs.zip"
SKILL_URL = "https://wx.gtimg.com/resource/wechatpay_api/test/skill/wechatpay-payment-integration/bundle.zip"
DOCS_TARGET_DIR = SKILL_DIR / "assets"
STATE_FILE = SCRIPT_DIR / ".wechatpay-resource-sync-state.json"
CHECK_INTERVAL_HOURS = 12

_TZ = timezone(timedelta(hours=8))

# ==== HTTP header（统一小写） ====

# 代理/网关可能会改写 Header 字段名大小写；字段名本身大小写不敏感。
# 这里统一将"字段名"转为大写进行匹配，避免因大小写变化导致取值失败。
H_LAST_MODIFIED = "LAST-MODIFIED"
H_ETAG = "ETAG"
H_CONTENT_LENGTH = "CONTENT-LENGTH"

# ==== 状态文件 key ====

# 检查/更新时间由 Skill 与知识库共用；远程版本标记各自独立。
S_LAST_CHECK = "LAST_CHECK_TIME"
S_LAST_UPDATE = "LAST_UPDATE_TIME"
S_DOCS_REMOTE_MODIFIED = "DOCS_REMOTE_LAST_MODIFIED"
S_DOCS_REMOTE_ETAG = "DOCS_REMOTE_ETAG"
S_SKILL_REMOTE_MODIFIED = "SKILL_REMOTE_LAST_MODIFIED"
S_SKILL_REMOTE_ETAG = "SKILL_REMOTE_ETAG"


USER_AGENT = f"{Path(__file__).stem}/{VERSION}"
IGNORED_FILES = {".DS_Store", "Thumbs.db", "__MACOSX"}
DOCS_FILE_GLOB = "*.md"
ZIP_FALLBACK_ENCODING = "cp437"

# Skill 更新时需要保留的顶层目录（由文档同步独立管理）
SKILL_PRESERVE_DIRS = {DOCS_TARGET_DIR.name}


# ==== 状态管理 ====


def _now_iso() -> str:
    """返回当前时间的 ISO 8601 字符串（东八区）。"""
    return datetime.now(_TZ).isoformat(timespec="seconds")


def _load_state() -> dict:
    """从 STATE_FILE 读取同步状态，文件不存在则返回空 dict。"""
    if STATE_FILE.exists():
        state = json.loads(STATE_FILE.read_text(encoding="utf-8"))
        # 兼容：部分环境可能会改写 header value 的大小写；本地 state 统一按小写存取。
        for key in (
            S_DOCS_REMOTE_ETAG,
            S_DOCS_REMOTE_MODIFIED,
            S_SKILL_REMOTE_ETAG,
            S_SKILL_REMOTE_MODIFIED,
        ):
            if isinstance(state.get(key), str):
                state[key] = state[key].lower()
        return state
    return {}


def _save_state(state: dict) -> None:
    """将同步状态写入 STATE_FILE。"""
    STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
    STATE_FILE.write_text(
        json.dumps(state, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )


# ==== 远程资源 ====


def _within_interval(state: dict) -> bool:
    """判断距上次检查是否不足 CHECK_INTERVAL_HOURS 小时。"""
    ts = state.get(S_LAST_CHECK)
    if not ts:
        return False
    elapsed = datetime.now(_TZ) - datetime.fromisoformat(ts)
    return elapsed.total_seconds() < CHECK_INTERVAL_HOURS * 3600


def _remote_changed(
    state: dict, remote: dict, etag_state_key: str, modified_state_key: str
) -> bool:
    """比较远程 ETag / Last-Modified 与本地记录，判断是否有变化。"""
    r_etag = remote.get(H_ETAG)
    if r_etag and state.get(etag_state_key):
        return str(r_etag).lower() != str(state[etag_state_key]).lower()
    r_lm = remote.get(H_LAST_MODIFIED)
    if r_lm and state.get(modified_state_key):
        return str(r_lm).lower() != str(state[modified_state_key]).lower()
    return True  # 无法判断时视为有变化


def _head_url(url: str) -> dict:
    """HEAD 请求指定 URL，返回 last_modified / etag / content_length 或 error。"""
    req = Request(url, method="HEAD")
    req.add_header("User-Agent", USER_AGENT)
    try:
        with urlopen(req, timeout=15) as resp:
            headers = {k.upper(): v for k, v in resp.headers.items()}
            lm = headers.get(H_LAST_MODIFIED)
            etag = headers.get(H_ETAG)
            return {
                H_LAST_MODIFIED: lm.lower() if isinstance(lm, str) else lm,
                H_ETAG: etag.lower() if isinstance(etag, str) else etag,
                H_CONTENT_LENGTH: headers.get(H_CONTENT_LENGTH),
            }
    except (URLError, HTTPError) as exc:
        return {"error": str(exc)}


# ==== 下载与解压 ====


def _download_url(url: str, dest: Path, label: str = "下载") -> None:
    """从指定 URL 下载文件到 dest，显示下载进度。"""
    req = Request(url)
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
                    if not is_tty and pct % 10 != 0 and pct < 100:
                        continue
                    msg = (
                        f"  {label}: {done / 1048576:.1f} MB / "
                        f"{total / 1048576:.1f} MB ({pct}%)"
                    )
                    if is_tty:
                        print(f"\r{msg}", end="", flush=True)
                    else:
                        print(msg, flush=True)
                else:
                    mb = int(done / 1048576)
                    if is_tty:
                        print(
                            f"\r  {label}: {done / 1048576:.1f} MB",
                            end="",
                            flush=True,
                        )
                    elif mb > last_mb:
                        last_mb = mb
                        print(f"  {label}: {done / 1048576:.1f} MB", flush=True)
        print()
    if total is not None and done != total:
        raise RuntimeError(f"下载不完整（{done}/{total} 字节），请重试。")


def _fix_zip_name(name: str) -> str:
    """未置 UTF-8 标志位的 zip 成员名，按 cp437 还原后依次尝试 utf-8、gbk。"""
    try:
        raw = name.encode(ZIP_FALLBACK_ENCODING)
    except UnicodeEncodeError:
        return name
    for encoding in ("utf-8", "gbk"):
        try:
            return raw.decode(encoding)
        except UnicodeDecodeError:
            continue
    return name  # 两种编码都不成立时保留原名，不要中断解压


def _extract(archive: Path, dest: Path) -> None:
    """解压 zip 或 tar.gz 到 dest。若 zip 未标记 UTF-8 标志位，做编码修正。"""
    dest.mkdir(parents=True, exist_ok=True)
    if zipfile.is_zipfile(archive):
        with zipfile.ZipFile(archive) as zf:
            for info in zf.infolist():
                if not (info.flag_bits & 0x800):
                    info.filename = _fix_zip_name(info.filename)
                zf.extract(info, dest)
        return
    try:
        with tarfile.open(archive) as tf:
            try:
                tf.extractall(dest, filter="data")
            except TypeError:
                # filter 参数需 Python 3.11.4+；低版本自行校验成员路径再解压
                root = dest.resolve()
                for member in tf.getmembers():
                    # 链接的目标不受成员路径校验约束，且后续复制会跟随链接读到
                    # 链接目标的内容，因此一律拒绝
                    if member.issym() or member.islnk():
                        raise RuntimeError(
                            f"归档包含链接文件，已中止: {member.name}"
                        )
                    if not (dest / member.name).resolve().is_relative_to(root):
                        raise RuntimeError(
                            f"归档包含非法路径，已中止: {member.name}"
                        )
                tf.extractall(dest)
        return
    except tarfile.ReadError:
        # 仅「读不出 tar 结构」才说明格式不对；filter 拦截等错误照原样抛出
        pass
    raise RuntimeError(
        f"无法识别压缩格式，请确认下载链接是否为 zip 或 tar.gz 文件: {archive.name}"
    )


def _find_content_root(extract_dir: Path) -> Path:
    """若解压后只有单一顶层目录，则进入该目录作为实际内容根。"""
    items = [p for p in extract_dir.iterdir() if p.name not in IGNORED_FILES]
    if len(items) == 1 and items[0].is_dir():
        return items[0]
    return extract_dir


def _win_long_path(path: Path) -> str:
    """Windows 长路径前缀，绕过 MAX_PATH(260) 限制。"""
    resolved = str(path.resolve())
    if resolved.startswith("\\\\?\\"):
        return resolved
    if resolved.startswith("\\\\"):
        return "\\\\?\\UNC\\" + resolved[2:]
    return "\\\\?\\" + resolved


def _chmod_writable(path: Path) -> None:
    """在原有权限位上追加写权限；符号链接跳过，避免改动链接目标。"""
    if path.is_symlink():
        return
    try:
        os.chmod(path, os.stat(path).st_mode | stat.S_IWUSR)
    except OSError:
        pass


def _on_rm_error(func, path, _exc_info) -> None:
    """只读文件/目录删除失败时，先改权限再重试。"""
    # POSIX 下能否删除取决于父目录是否可写，父目录也要一并补权限
    _chmod_writable(Path(path).parent)
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


def _mkdir(path: Path) -> None:
    """创建目录；Windows 下失败时使用长路径重试。"""
    try:
        path.mkdir(parents=True, exist_ok=True)
        return
    except OSError:
        if sys.platform != "win32":
            raise
    os.makedirs(_win_long_path(path), exist_ok=True)


def _copy_file(src: Path, dst: Path) -> None:
    """复制单个文件；Windows 下自动尝试长路径。"""
    _mkdir(dst.parent)
    try:
        shutil.copy2(src, dst)
        return
    except OSError:
        if sys.platform != "win32":
            raise
    shutil.copy2(_win_long_path(src), _win_long_path(dst))


def _copy_tree(
    src: Path, dst: Path, preserve_top_level=None
) -> list[tuple[Path, str]]:
    """逐文件复制目录树，可跳过指定顶层目录，返回复制错误列表。"""
    preserve_top_level = preserve_top_level or set()
    _mkdir(dst)
    errors: list[tuple[Path, str]] = []
    for item in sorted(src.rglob("*")):
        rel = item.relative_to(src)
        if any(part in IGNORED_FILES for part in rel.parts):
            continue
        if rel.parts and rel.parts[0] in preserve_top_level:
            continue
        target = dst / rel
        if item.is_dir():
            try:
                _mkdir(target)
            except OSError as exc:
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
    if not resolved.is_relative_to(skill_root):
        raise RuntimeError(f"目标目录不在 skill 范围内，拒绝清空: {DOCS_TARGET_DIR}")
    for child in DOCS_TARGET_DIR.iterdir():
        if child.is_symlink():
            _unlink(child)
        elif child.is_dir():
            _remove_tree(child)
        else:
            _unlink(child)


# ==== Skill 目录管理 ====


def _clear_skill_dir() -> None:
    """清空 SKILL_DIR 下的内容，保留 assets 目录和状态文件。"""
    if not SKILL_DIR.exists():
        return
    state_resolved = STATE_FILE.resolve()
    assets_resolved = DOCS_TARGET_DIR.resolve()

    for child in list(SKILL_DIR.iterdir()):
        child_resolved = child.resolve()
        if child_resolved == assets_resolved:
            continue
        if not child.is_dir():
            if child_resolved != state_resolved:
                _unlink(child)
            continue
        # 检查状态文件是否在此目录下
        _state_inside = False
        try:
            state_resolved.relative_to(child_resolved)
            _state_inside = True
        except ValueError:
            pass
        if not _state_inside:
            _remove_tree(child)
            continue
        # 状态文件在此目录内，逐项清理但保留状态文件本身
        for sub in list(child.iterdir()):
            if sub.resolve() == state_resolved:
                continue
            if sub.is_dir():
                _remove_tree(sub)
            else:
                _unlink(sub)


# ==== 命令 ====


def _is_installed() -> bool:
    """判断本地知识库目录是否存在且非空。"""
    return DOCS_TARGET_DIR.exists() and any(DOCS_TARGET_DIR.iterdir())


def cmd_update(state: dict, skip_check: bool) -> None:
    """检查远程版本；有更新或首次安装时下载、清空 assets 并写入新版本。"""
    installed = _is_installed()

    if installed and skip_check:
        last = state.get(S_LAST_CHECK, "未知")
        print(
            f"知识库已是最新（上次检查: {last}，{CHECK_INTERVAL_HOURS}H 内无需重复检查）。"
        )
        return

    print("正在检查远程文档版本…")
    remote = _head_url(DOCS_URL)
    if "error" in remote:
        print(
            f"无法连接远程服务器，请检查网络后重试。\n  错误详情: {remote['error']}",
            file=sys.stderr,
        )
        sys.exit(1)

    state[S_LAST_CHECK] = _now_iso()

    if installed and not _remote_changed(
        state, remote, S_DOCS_REMOTE_ETAG, S_DOCS_REMOTE_MODIFIED
    ):
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

        _download_url(DOCS_URL, archive)

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
    state[S_DOCS_REMOTE_MODIFIED] = str(remote.get(H_LAST_MODIFIED, "") or "").lower()
    state[S_DOCS_REMOTE_ETAG] = str(remote.get(H_ETAG, "") or "").lower()
    _save_state(state)

    count = sum(1 for _ in DOCS_TARGET_DIR.rglob(DOCS_FILE_GLOB))
    print(f"更新完成，当前共 {count} 篇文档。")


def _is_skill_installed() -> bool:
    """判断 skill 是否已安装（SKILL.md 存在即视为已安装）。"""
    return (SKILL_DIR / "SKILL.md").exists()


def cmd_skill_update(state: dict, skip_check: bool) -> None:
    """检查 Skill 远程版本；有更新时下载并替换（保留 assets 目录和状态文件）。"""
    installed = _is_skill_installed()

    if installed and skip_check:
        last = state.get(S_LAST_CHECK, "未知")
        print(
            f"Skill 已是最新（上次检查: {last}，{CHECK_INTERVAL_HOURS}H 内无需重复检查）。"
        )
        return

    print("正在检查 Skill 远程版本…")
    remote = _head_url(SKILL_URL)
    if "error" in remote:
        print(
            f"无法连接 Skill 远程服务器，请检查网络后重试。\n"
            f"  错误详情: {remote['error']}",
            file=sys.stderr,
        )
        sys.exit(1)

    state[S_LAST_CHECK] = _now_iso()

    if installed and not _remote_changed(
        state, remote, S_SKILL_REMOTE_ETAG, S_SKILL_REMOTE_MODIFIED
    ):
        _save_state(state)
        print("Skill 远程未发生变化，当前已是最新，无需更新。")
        return

    if not installed:
        print("本地尚未安装 Skill，开始首次下载…")
    else:
        print(
            f"检测到 Skill 已更新（{remote.get(H_LAST_MODIFIED, '时间未知')}），"
            f"开始下载…"
        )

    with tempfile.TemporaryDirectory() as tmp:
        tmp = Path(tmp)
        archive = tmp / "bundle.tar.gz"

        _download_url(SKILL_URL, archive, label="下载 Skill")

        print("下载完成，正在解压…")
        extract_dir = tmp / "out"
        _extract(archive, extract_dir)
        new_root = _find_content_root(extract_dir)

        print("正在更新 Skill…")
        _clear_skill_dir()
        copy_errors = _copy_tree(new_root, SKILL_DIR, SKILL_PRESERVE_DIRS)
        if copy_errors:
            print(
                f"更新 Skill 时出现 {len(copy_errors)} 个文件错误：",
                file=sys.stderr,
            )
            for rel, msg in copy_errors[:10]:
                print(f"  - {rel}: {msg}", file=sys.stderr)
            if len(copy_errors) > 10:
                print(
                    f"  … 另有 {len(copy_errors) - 10} 个错误未列出", file=sys.stderr
                )
            sys.exit(1)

    state[S_LAST_UPDATE] = _now_iso()
    state[S_SKILL_REMOTE_MODIFIED] = str(
        remote.get(H_LAST_MODIFIED, "") or ""
    ).lower()
    state[S_SKILL_REMOTE_ETAG] = str(remote.get(H_ETAG, "") or "").lower()
    _save_state(state)

    print("Skill 更新完成。")


# ==== 入口 ====
_USAGE = """\
用法: python3 wechatpay-resource-sync.py update

  update   检查远程 Skill 和知识库是否有更新；有变化时下载并替换（含首次安装）
           默认 12 小时内不重复检查远程"""


def main() -> None:
    if len(sys.argv) < 2 or sys.argv[1] in ("-h", "--help"):
        print(_USAGE)
        sys.exit(0)

    cmd = sys.argv[1]

    if cmd != "update":
        print(f"未识别到有效命令: {cmd}\n")
        print(_USAGE)
        sys.exit(1)

    # 状态只读一次；间隔判断必须在任何落盘之前算好，否则 Skill 更新写入的
    # LAST_CHECK_TIME 会让紧随其后的知识库检查被误判为“12H 内已检查”。
    state = _load_state()
    skip_check = _within_interval(state)

    # Skill 与知识库共用同一个 LAST_CHECK_TIME，两者都在间隔内时时间戳必然相同，
    # 合并成一条提示，避免打印两遍。
    if skip_check and _is_skill_installed() and _is_installed():
        last = state.get(S_LAST_CHECK, "未知")
        print(
            f"Skill 与知识库均已是最新（上次检查: {last}，"
            f"{CHECK_INTERVAL_HOURS}H 内无需重复检查）。"
        )
        return

    cmd_skill_update(state, skip_check)
    cmd_update(state, skip_check)


if __name__ == "__main__":
    main()
