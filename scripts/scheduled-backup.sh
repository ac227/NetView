#!/usr/bin/env bash
# 定时备份任务（供 launchd 调用）：每天备份，保留 30 份
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec "$ROOT/scripts/backup.sh" "$ROOT/backups" 30
