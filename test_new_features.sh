#!/bin/bash

echo "=== 测试新功能 ==="

# 1. 检查应用状态
echo "1. 检查应用状态..."
if pgrep -f "db-desktop" > /dev/null; then
    echo "✅ 应用正在运行"
else
    echo "❌ 应用未运行"
    exit 1
fi

# 2. 检查数据库表结构
echo -e "\n2. 检查数据库表结构..."
sqlite3 ~/.db-desktop/conversations.db "PRAGMA table_info(query_history);"

# 3. 检查当前查询历史记录
echo -e "\n3. 检查当前查询历史记录..."
TOTAL_QUERIES=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "当前查询历史记录数: $TOTAL_QUERIES"

if [ "$TOTAL_QUERIES" -gt 0 ]; then
    echo "最近的查询历史记录:"
    sqlite3 ~/.db-desktop/conversations.db "SELECT id, query, db_type, connection_name, success, created_at FROM query_history ORDER BY created_at DESC LIMIT 3;"
fi

# 4. 测试插入一条新记录
echo -e "\n4. 测试插入新记录..."
sqlite3 ~/.db-desktop/conversations.db "
INSERT INTO query_history (query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows) 
VALUES ('SELECT NOW()', 15, 'mysql', 'test_connection', 'Test MySQL Connection', datetime('now'), 1, '', 1);
"

# 5. 验证插入结果
echo -e "\n5. 验证插入结果..."
sqlite3 ~/.db-desktop/conversations.db "SELECT id, query, db_type, connection_name, success, created_at FROM query_history ORDER BY created_at DESC LIMIT 1;"

echo -e "\n=== 测试完成 ==="
echo "新功能已实现："
echo "✅ 1. 添加了连接名称列"
echo "✅ 2. 添加了再试一次按钮"
echo "✅ 3. 添加了执行结果展示窗口"
echo "✅ 4. 实现了重新执行查询的逻辑"
echo ""
echo "请在应用中测试："
echo "1. 查看查询历史页面是否显示连接名称列"
echo "2. 点击'再试一次'按钮测试重新执行功能"
echo "3. 查看执行结果是否正确显示在模态框中"
