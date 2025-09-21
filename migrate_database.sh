#!/bin/bash

echo "=== 数据库迁移脚本 ==="

# 检查SQLite数据库是否存在
if [ ! -f ~/.db-desktop/conversations.db ]; then
    echo "❌ SQLite数据库不存在，无需迁移"
    exit 0
fi

echo "✅ 找到SQLite数据库"

# 检查是否已经存在新列
echo "检查表结构..."
HAS_CONNECTION_ID=$(sqlite3 ~/.db-desktop/conversations.db "PRAGMA table_info(query_history);" | grep "connection_id" | wc -l)
HAS_CONNECTION_NAME=$(sqlite3 ~/.db-desktop/conversations.db "PRAGMA table_info(query_history);" | grep "connection_name" | wc -l)

if [ "$HAS_CONNECTION_ID" -gt 0 ] && [ "$HAS_CONNECTION_NAME" -gt 0 ]; then
    echo "✅ 表结构已经是最新的，无需迁移"
    exit 0
fi

echo "🔄 开始迁移表结构..."

# 备份原表
echo "备份原表..."
sqlite3 ~/.db-desktop/conversations.db "
CREATE TABLE query_history_backup AS SELECT * FROM query_history;
"

# 删除原表
echo "删除原表..."
sqlite3 ~/.db-desktop/conversations.db "
DROP TABLE query_history;
"

# 创建新表结构
echo "创建新表结构..."
sqlite3 ~/.db-desktop/conversations.db "
CREATE TABLE query_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT NOT NULL,
    execution_time INTEGER NOT NULL,
    db_type TEXT NOT NULL,
    connection_id TEXT NOT NULL DEFAULT 'unknown',
    connection_name TEXT NOT NULL DEFAULT 'Unknown Connection',
    created_at DATETIME NOT NULL,
    success BOOLEAN NOT NULL DEFAULT 1,
    error TEXT,
    result_rows INTEGER DEFAULT 0
);
"

# 创建索引
echo "创建索引..."
sqlite3 ~/.db-desktop/conversations.db "
CREATE INDEX IF NOT EXISTS idx_query_history_created_at ON query_history (created_at);
CREATE INDEX IF NOT EXISTS idx_query_history_db_type ON query_history (db_type);
CREATE INDEX IF NOT EXISTS idx_query_history_success ON query_history (success);
CREATE INDEX IF NOT EXISTS idx_query_history_connection_id ON query_history (connection_id);
CREATE INDEX IF NOT EXISTS idx_query_history_connection_name ON query_history (connection_name);
"

# 迁移数据
echo "迁移数据..."
sqlite3 ~/.db-desktop/conversations.db "
INSERT INTO query_history (id, query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows)
SELECT id, query, execution_time, db_type, 'unknown', 'Unknown Connection', created_at, success, error, result_rows
FROM query_history_backup;
"

# 删除备份表
echo "清理备份表..."
sqlite3 ~/.db-desktop/conversations.db "
DROP TABLE query_history_backup;
"

echo "✅ 数据库迁移完成"

# 验证迁移结果
echo "验证迁移结果..."
TOTAL_RECORDS=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "迁移后记录数: $TOTAL_RECORDS"

echo "=== 迁移完成 ==="
