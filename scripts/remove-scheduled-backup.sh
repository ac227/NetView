#!/usr/bin/env bash
# 卸载 macOS 定时备份
set -euo pipefail
LABEL="com.netview.backup"
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"

if [ -f "$PLIST" ]; then
  launchctl unload "$PLIST" 2>/dev/null || true
  rm -f "$PLIST"
  echo "已移除定时备份任务：$PLIST"
else
  echo "没有已安装的定时备份任务。"
fi
