#!/usr/bin/env bash
# 安装 macOS 定时备份（launchd，每天 03:00 自动备份，保留 30 份）
# 用法：./scripts/setup-scheduled-backup.sh [HH:MM]   （默认 03:00）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

LABEL="com.netview.backup"
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"
TIME="${1:-03:00}"
HOUR="${TIME%%:*}"
MINUTE="${TIME##*:}"
LOG_DIR="$ROOT/backups"
JOB="$ROOT/scripts/scheduled-backup.sh"

mkdir -p "$LOG_DIR"

cat > "$PLIST" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$LABEL</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>$JOB</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>$HOUR</integer>
        <key>Minute</key>
        <integer>$MINUTE</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/backup.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/backup.log</string>
</dict>
</plist>
EOF

launchctl unload "$PLIST" 2>/dev/null || true
launchctl load "$PLIST"
echo "已安装定时备份：每天 $TIME 执行，保留 30 份"
echo "  任务配置：$PLIST"
echo "  备份目录：$LOG_DIR"
echo "  日志：    $LOG_DIR/backup.log"
echo "卸载：./scripts/remove-scheduled-backup.sh"
