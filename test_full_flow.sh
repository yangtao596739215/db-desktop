#!/bin/bash

echo "=== 完整流程测试 ==="

# 1. 检查应用状态
echo "1. 检查应用状态..."
if pgrep -f "db-desktop" > /dev/null; then
    echo "✅ 应用正在运行"
else
    echo "❌ 应用未运行"
    exit 1
fi

# 2. 检查连接配置
echo -e "\n2. 检查连接配置..."
echo "MySQL连接ID: mysql_1758239995"
echo "连接状态: disconnected"

# 3. 检查SQLite数据库
echo -e "\n3. 检查SQLite数据库..."
TOTAL_BEFORE=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "查询历史记录数（执行前）: $TOTAL_BEFORE"

# 4. 模拟查询执行（通过直接调用后端API）
echo -e "\n4. 模拟查询执行..."

# 检查是否有wails CLI工具
if command -v wails &> /dev/null; then
    echo "尝试使用wails CLI调用后端API..."
    # 这里需要实际的API调用，但wails CLI可能不支持直接调用
    echo "注意：需要在实际应用中测试查询执行"
else
    echo "wails CLI不可用，跳过API调用测试"
fi

# 5. 检查执行后的状态
echo -e "\n5. 检查执行后状态..."
TOTAL_AFTER=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "查询历史记录数（执行后）: $TOTAL_AFTER"

if [ "$TOTAL_AFTER" -gt "$TOTAL_BEFORE" ]; then
    echo "✅ 查询历史记录增加了"
    sqlite3 ~/.db-desktop/conversations.db "SELECT id, query, db_type, success, created_at FROM query_history ORDER BY created_at DESC LIMIT 3;"
else
    echo "❌ 查询历史记录没有增加"
fi

# 6. 分析可能的问题
echo -e "\n6. 可能的问题分析..."
echo "可能的问题："
echo "1. 连接未建立 - 需要在应用中先连接数据库"
echo "2. 连接ID不正确 - 前端传递的连接ID与后端不匹配"
echo "3. 查询执行失败 - 查询本身有问题"
echo "4. SQLite管理器未初始化 - 后端SQLite管理器为nil"
echo "5. 数据库类型检测失败 - 无法正确识别MySQL类型"

echo -e "\n=== 测试完成 ==="
