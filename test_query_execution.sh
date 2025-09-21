#!/bin/bash

echo "=== 测试查询执行和历史保存 ==="

# 检查应用是否运行
if ! pgrep -f "db-desktop" > /dev/null; then
    echo "❌ 应用未运行，请先启动应用"
    exit 1
fi

echo "✅ 应用正在运行"

# 检查连接配置
echo -e "\n1. 检查连接配置..."
if [ -f ~/.db-desktop/connections.json ]; then
    echo "连接配置文件内容:"
    cat ~/.db-desktop/connections.json | jq .
else
    echo "❌ 连接配置文件不存在"
    exit 1
fi

# 检查SQLite数据库状态
echo -e "\n2. 检查SQLite数据库状态..."
sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) as total_queries FROM query_history;"

# 模拟插入一条测试记录
echo -e "\n3. 手动插入测试查询历史记录..."
sqlite3 ~/.db-desktop/conversations.db "
INSERT INTO query_history (query, execution_time, db_type, created_at, success, error, result_rows) 
VALUES ('SELECT 1', 10, 'mysql', datetime('now'), 1, '', 1);
"

# 验证插入是否成功
echo -e "\n4. 验证插入结果..."
sqlite3 ~/.db-desktop/conversations.db "SELECT * FROM query_history ORDER BY created_at DESC LIMIT 1;"

echo -e "\n=== 测试完成 ==="
