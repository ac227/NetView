#!/usr/bin/env bash
# NetView 完整备份：数据库 + 媒体文件 + 配置，带保留策略与校验
# 用法：./scripts/backup.sh [输出目录] [保留份数]
#   输出目录    默认 ./backups
#   保留份数    默认 10，备份超过该数量时自动删除最旧的
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

BACKUP_DIR="${1:-$ROOT/backups}"
KEEP="${2:-10}"
DATA_DIR="${NETVIEW_DATA_DIR:-$ROOT/bin/data}"
DB_DSN="${NETVIEW_DB_DSN:-postgres://netview:netview_dev@localhost:5432/netview}"
STAMP="$(date +%Y%m%d_%H%M%S)"
OUT="$BACKUP_DIR/netview_$STAMP"

mkdir -p "$OUT"

echo "==> 1/4 备份数据库（custom 格式，含设置/密码哈希/标签/分类/下载任务）"
DSN_NO_SCHEME="${DB_DSN#postgres://}"
USER_PART="${DSN_NO_SCHEME%%@*}"
HOST_DB="${DSN_NO_SCHEME#*@}"
HOST_PART="${HOST_DB%%/*}"
DB_NAME="${HOST_DB#*/}"
export PGPASSWORD="${USER_PART#*:}"
pg_dump -h "${HOST_PART%%:*}" -p "${HOST_PART##*:}" -U "${USER_PART%%:*}" -F c -f "$OUT/database.dump" "$DB_NAME"

echo "==> 2/4 校验数据库备份"
pg_restore -l "$OUT/database.dump" > /dev/null 2>&1 \
  && echo "    database.dump 校验通过（可被 pg_restore 读取）" \
  || { echo "    [警告] database.dump 校验失败"; }

echo "==> 3/4 打包媒体文件"
if [ -d "$DATA_DIR" ] && [ -n "$(ls -A "$DATA_DIR" 2>/dev/null)" ]; then
  tar -czf "$OUT/media.tar.gz" -C "$DATA_DIR" .
else
  echo "    数据目录为空，跳过媒体打包"
  touch "$OUT/media.tar.gz"
fi
tar -tzf "$OUT/media.tar.gz" > /dev/null 2>&1 \
  && echo "    media.tar.gz 校验通过（可正常读取）" \
  || echo "    [警告] media.tar.gz 校验失败"

echo "==> 4/4 附带配置文件（如有 .env）"
if [ -f "$ROOT/.env" ]; then
  cp "$ROOT/.env" "$OUT/env.txt"
fi

# 保留策略：删除最旧的备份，只保留最近 KEEP 份
ls -1dt "$BACKUP_DIR"/netview_* 2>/dev/null | tail -n +$((KEEP + 1)) | while read -r old; do
  echo "==> 删除过期备份：$(basename "$old")"
  rm -rf "$old"
done

echo ""
echo "完成：$OUT"
echo "    数据量：$(du -sh "$OUT" | cut -f1)"
echo "    恢复：./scripts/restore.sh $OUT"
