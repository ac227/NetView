#!/usr/bin/env bash
# 启动 NetView：直接运行编译好的二进制（无需容器、无需 Node）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin/netview"

if [ ! -f "$BIN" ]; then
  echo "未找到 $BIN"
  echo "请先执行 ./scripts/build.sh 编译一次。"
  exit 1
fi

exec "$BIN" "$@"
