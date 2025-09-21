#!/bin/bash

echo "=== 测试查询历史保存功能 ==="

# 检查SQLite数据库
echo "1. 检查SQLite数据库状态..."
if [ -f ~/.db-desktop/conversations.db ]; then
    echo "✅ SQLite数据库文件存在"
    echo "数据库大小: $(ls -lh ~/.db-desktop/conversations.db | awk '{print $5}')"
else
    echo "❌ SQLite数据库文件不存在"
    exit 1
fi

# 检查表结构
echo -e "\n2. 检查查询历史表结构..."
sqlite3 ~/.db-desktop/conversations.db "PRAGMA table_info(query_history);"

# 检查当前记录数
echo -e "\n3. 检查当前查询历史记录数..."
TOTAL_QUERIES=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "当前查询历史记录数: $TOTAL_QUERIES"

# 检查最近的记录
echo -e "\n4. 检查最近的查询历史记录..."
sqlite3 ~/.db-desktop/conversations.db "SELECT id, query, db_type, success, created_at FROM query_history ORDER BY created_at DESC LIMIT 5;"

# 检查连接表
echo -e "\n5. 检查连接配置..."
sqlite3 ~/.db-desktop/conversations.db "SELECT name FROM sqlite_master WHERE type='table';"

# 检查应用进程
echo -e "\n6. 检查应用进程状态..."
ps aux | grep "db-desktop" | grep -v grep

echo -e "\n=== 测试完成 ==="
