#!/usr/bin/env bash
# 初始化本地 PostgreSQL 数据库（Homebrew 版）
set -euo pipefail

DB_USER="${NETVIEW_DB_USER:-netview}"
DB_PASS="${NETVIEW_DB_PASS:-netview_dev}"
DB_NAME="${NETVIEW_DB_NAME:-netview}"

if ! command -v psql >/dev/null 2>&1; then
  echo "未找到 psql。请先安装 PostgreSQL 16，例如：brew install postgresql@16 && brew services start postgresql@16"
  exit 1
fi

psql -d postgres -v ON_ERROR_STOP=1 <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASS}';
  END IF;
END
\$\$;
SQL

psql -d postgres -v ON_ERROR_STOP=1 \
  -c "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'" | grep -q 1 || \
  createdb -O "${DB_USER}" "${DB_NAME}"

echo "数据库 ${DB_NAME} 已就绪（用户 ${DB_USER}）。"
echo "连接串: postgres://${DB_USER}:${DB_PASS}@localhost:5432/${DB_NAME}"
