#!/usr/bin/env bash
# 恢复 NetView 备份
# 用法：./scripts/restore.sh <备份目录>   （如 backups/netview_20260809_130000）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

BACKUP_DIR="${1:?用法: ./scripts/restore.sh <备份目录>}"
DB_DSN="${NETVIEW_DB_DSN:-postgres://netview:netview_dev@localhost:5432/netview}"
DATA_DIR="${NETVIEW_DATA_DIR:-$ROOT/bin/data}"

[ -f "$BACKUP_DIR/database.dump" ] || { echo "缺少 database.dump"; exit 1; }
[ -f "$BACKUP_DIR/media.tar.gz" ] || { echo "缺少 media.tar.gz"; exit 1; }

echo "==> 1/2 恢复数据库（将清空并重建现有库）"
DSN_NO_SCHEME="${DB_DSN#postgres://}"
USER_PART="${DSN_NO_SCHEME%%@*}"
HOST_DB="${DSN_NO_SCHEME#*@}"
HOST_PART="${HOST_DB%%/*}"
DB_NAME="${HOST_DB#*/}"
export PGPASSWORD="${USER_PART#*:}"
pg_restore --clean --if-exists -h "${HOST_PART%%:*}" -p "${HOST_PART##*:}" -U "${USER_PART%%:*}" -d "$DB_NAME" "$BACKUP_DIR/database.dump"

echo "==> 2/2 解压媒体文件"
mkdir -p "$DATA_DIR"
tar -xzf "$BACKUP_DIR/media.tar.gz" -C "$DATA_DIR"

echo "恢复完成。"
