#!/usr/bin/env bash
# 开发模式：同时启动后端(8080)和前端(5173)，前端通过 /api 代理到后端
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

trap 'kill 0' EXIT

# 开发模式下固定数据目录，避免 go run 临时目录导致数据漂移
export NETVIEW_DATA_DIR="${NETVIEW_DATA_DIR:-$ROOT/data}"

echo "==> 启动后端 (http://localhost:8080)"
(cd "$ROOT/backend" && go run ./cmd/server) &

echo "==> 启动前端 (http://localhost:5173)"
(cd "$ROOT/frontend" && npm run dev) &

wait
