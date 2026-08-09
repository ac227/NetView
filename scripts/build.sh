#!/usr/bin/env bash
# 构建单文件可执行程序：编译前端并嵌入 Go 二进制，产物为 bin/netview
# 用法：./scripts/build.sh [版本号]   （版本号默认 dev，会在 -version 输出）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_DIR="$ROOT/bin"
WEB_DIST="$ROOT/backend/internal/web/dist"
VERSION="${1:-dev}"

echo "==> 1/3 构建前端"
(cd "$ROOT/frontend" && npm ci && npm run build)

echo "==> 2/3 复制前端产物到后端（嵌入二进制）"
rm -rf "$WEB_DIST"
cp -r "$ROOT/frontend/dist" "$WEB_DIST"
touch "$WEB_DIST/.gitkeep"

echo "==> 3/3 编译二进制（版本 ${VERSION}）"
mkdir -p "$BIN_DIR"
(cd "$ROOT/backend" && CGO_ENABLED=0 go build -trimpath \
  -ldflags="-s -w -X main.version=$VERSION" \
  -o "$BIN_DIR/netview" ./cmd/server)

echo ""
echo "完成：$BIN_DIR/netview (NetView $VERSION)"
echo "运行：$BIN_DIR/netview"
echo "访问：http://localhost:8080"
