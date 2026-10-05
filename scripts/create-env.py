#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""以 .env.example 為範本,互動式產生 .env.

只詢問第一次啟動必須確認的欄位,其餘內容(含註解)原封不動從 .env.example 複製.

預設值優先順序: 既有 .env > .env.example > 腳本內建值.
SYSTEM_AES_KEY / JWT_SECRET 不採用 .env.example 的值(範本是公開的),
留空時自動以亂數產生;若既有 .env 已有合法值則沿用,避免已加密資料解不開.

Usage:
    uv run scripts/create-env.py          # 互動式填寫
    uv run scripts/create-env.py -y       # 全部採用預設值
    uv run scripts/create-env.py --example path/.env.example --output path/.env
"""

import argparse
import re
import secrets
import shutil
import sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# 只比對大寫變數名,避免把註解裡的說明文字(例如 "# default=xxx")當成設定.
LINE_RE = re.compile(r"^(?P<indent>\s*)(?P<comment>#\s*)?(?P<key>[A-Z_][A-Z0-9_]*)\s*=(?P<rest>.*)$")

# 不需加引號的字元;其他字元一律用單引號包起來,docker compose 和 bash source 都不會展開.
SAFE_VALUE_RE = re.compile(r"^[A-Za-z0-9_\-.:/@+=,%]*$")
PLACEHOLDER_RE = re.compile(r"(^your[-_]|change[-_]?me|^x{3,}$)", re.IGNORECASE)

# 與後端黑名單一致: internal/utils/presign.go, internal/application/service/user.go
KNOWN_EXAMPLE_KEYS = {
    "weknora-system-aes-key-32bytes!!",
    "your-32-byte-long-encryption-key!",
    "weknora-jwt-secret",
    "CHANGE-ME-jwt-secret",
}


@dataclass(frozen=True)
class Field:
    name: str
    label: str
    fallback: str = ""
    kind: str = "text"  # text | port | aes_key | secret


FIELDS = [
    Field("DB_USER", "PostgreSQL 使用者", "weknora"),
    Field("DB_PASSWORD", "PostgreSQL 密碼", "weknora_db_password"),
    Field("DB_NAME", "PostgreSQL 資料庫名稱", "WeKnora"),
    Field("REDIS_PASSWORD", "Redis 密碼", "weknora_redis_password"),
    Field("SYSTEM_AES_KEY", "SYSTEM_AES_KEY (必須剛好 32 bytes)", kind="aes_key"),
    Field("JWT_SECRET", "JWT_SECRET (至少 32 字元)", kind="secret"),
    Field("FRONTEND_PORT", "前端對外 port", "80", kind="port"),
    Field("APP_PORT", "後端 API 對外 port", "8080", kind="port"),
    Field("TZ", "時區", "Asia/Taipei"),
    Field("OLLAMA_BASE_URL", "Ollama 位址", "http://host.docker.internal:11434"),
]


def parse_value(rest):
    """把等號右側拆成 (值, 行尾註解)."""
    rest = rest.strip()
    if rest[:1] in ("'", '"'):
        end = rest.find(rest[0], 1)
        if end > 0:
            return rest[1:end], rest[end + 1 :]
    match = re.search(r"\s+#", rest)
    if match:
        return rest[: match.start()].strip(), rest[match.start() :]
    return rest, ""


def read_env(path):
    """讀取 env 檔的有效設定(忽略被註解的行),同名時後者覆蓋前者."""
    values = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = LINE_RE.match(line)
        if match and match["comment"] is None:
            values[match["key"]] = parse_value(match["rest"])[0]
    return values


def validate(field, value):
    """回傳錯誤訊息;合法時回傳 None."""
    if re.search(r"[\s'\"`\\]", value):
        return "不可包含空白、引號或反斜線"
    if field.kind == "port":
        if not value.isdigit() or not 1 <= int(value) <= 65535:
            return "port 必須是 1-65535 的整數"
    elif field.kind in ("aes_key", "secret"):
        if value in KNOWN_EXAMPLE_KEYS:
            return "這是公開的範例值,後端會拒絕使用"
        if field.kind == "aes_key" and len(value.encode()) != 32:
            return f"長度必須剛好 32 bytes (目前 {len(value.encode())})"
        if field.kind == "secret" and len(value) < 32:
            return f"長度至少 32 字元 (目前 {len(value)})"
    elif not value:
        return "不可為空"
    return None


def generate_key(field):
    # token_hex 只會產生 [0-9a-f],32 字元剛好等於 32 bytes.
    return secrets.token_hex(16) if field.kind == "aes_key" else secrets.token_hex(32)


def resolve_default(field, existing, example):
    """依優先順序找出第一個合法且非 placeholder 的值."""
    is_key = field.kind in ("aes_key", "secret")
    sources = [existing] if is_key else [existing, example]
    for source in sources:
        value = source.get(field.name, "")
        if value and not PLACEHOLDER_RE.search(value) and validate(field, value) is None:
            return value
    return "" if is_key else field.fallback


def mask(value):
    return value[:4] + "*" * 8


def ask(field, default, assume_yes):
    is_key = field.kind in ("aes_key", "secret")
    if is_key:
        shown = f"沿用現有值 {mask(default)}" if default else "留空自動產生"
    else:
        shown = default

    while True:
        value = default if assume_yes else input(f"{field.label} [{shown}]: ").strip() or default
        if is_key and not value:
            return generate_key(field)
        error = validate(field, value)
        if error is None:
            return value
        if assume_yes:
            sys.exit(f"[錯誤] {field.name} 的預設值不合法: {error}")
        print(f"  -> {error}, 請重新輸入.")


def format_value(value):
    return value if SAFE_VALUE_RE.match(value) else f"'{value}'"


def render(example_lines, values):
    """以範本為基礎替換欄位值;只有被註解的欄位會取消註解,範本沒有的欄位附加在檔尾."""
    active_keys = {
        m["key"] for m in map(LINE_RE.match, example_lines) if m and m["comment"] is None
    }
    written = set()
    out = []
    for line in example_lines:
        match = LINE_RE.match(line)
        if match and match["key"] in values:
            key = match["key"]
            is_active = match["comment"] is None
            if is_active or (key not in active_keys and key not in written):
                trailing = parse_value(match["rest"])[1]
                out.append(f"{match['indent']}{key}={format_value(values[key])}{trailing}")
                written.add(key)
                continue
        out.append(line)

    missing = [key for key in values if key not in written]
    if missing:
        out += ["", "# ===== 由 scripts/create-env.py 新增 ====="]
        out += [f"{key}={format_value(values[key])}" for key in missing]
    return "\n".join(out) + "\n"


def confirm(question):
    return input(f"{question} [y/N]: ").strip().lower() in ("y", "yes")


def main():
    parser = argparse.ArgumentParser(description="以 .env.example 為範本產生 .env")
    parser.add_argument("--example", type=Path, default=ROOT / ".env.example")
    parser.add_argument("--output", type=Path, default=ROOT / ".env")
    parser.add_argument("-y", "--yes", action="store_true", help="不詢問,全部採用預設值")
    args = parser.parse_args()

    if not args.example.is_file():
        sys.exit(f"[錯誤] 找不到範本 {args.example}")

    example_lines = args.example.read_text(encoding="utf-8").splitlines()
    example = read_env(args.example)
    existing = {}
    if args.output.exists():
        print(f"{args.output} 已存在,現有值會當作預設值,原檔會先備份.")
        if not args.yes and not confirm("要繼續覆寫嗎?"):
            print("已取消.")
            return
        existing = read_env(args.output)

    print("直接按 Enter 採用 [] 內的預設值.\n")
    values = {f.name: ask(f, resolve_default(f, existing, example), args.yes) for f in FIELDS}

    if args.output.exists():
        backup = args.output.with_name(f"{args.output.name}.bak-{datetime.now():%Y%m%d%H%M%S}")
        shutil.copy2(args.output, backup)
        print(f"\n已備份原檔至 {backup}")

    # 固定 LF 換行: start_all.sh 會用 bash source .env,CRLF 會讓值多出 \r.
    args.output.write_text(render(example_lines, values), encoding="utf-8", newline="\n")

    print(f"\n已寫入 {args.output}:")
    for f in FIELDS:
        value = values[f.name]
        shown = mask(value) if f.kind in ("aes_key", "secret") else value
        print(f"  {f.name}={shown}")
    if existing.get("SYSTEM_AES_KEY") and existing["SYSTEM_AES_KEY"] != values["SYSTEM_AES_KEY"]:
        print("\n[注意] SYSTEM_AES_KEY 已變更,DB 中已加密的模型 API key 將無法解密,需在 UI 重新輸入.")


if __name__ == "__main__":
    try:
        main()
    except (KeyboardInterrupt, EOFError):
        print("\n已取消.")
        sys.exit(130)
