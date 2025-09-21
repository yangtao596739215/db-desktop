#!/bin/bash

echo "=== 验证修复效果 ==="

# 1. 重新编译应用
echo "1. 重新编译应用..."
cd /Users/yangtao/WorkProject/personal/db-desktop

# 停止当前应用
echo "停止当前应用..."
pkill -f "db-desktop" || true
sleep 2

# 编译应用
echo "编译应用..."
if command -v wails &> /dev/null; then
    wails build
    if [ $? -eq 0 ]; then
        echo "✅ 编译成功"
    else
        echo "❌ 编译失败"
        exit 1
    fi
else
    echo "wails命令不可用，跳过编译"
fi

# 2. 启动应用
echo -e "\n2. 启动应用..."
if [ -f "./build/bin/db-desktop.app/Contents/MacOS/A modern database management tool for MySQL, Redis, and ClickHouse" ]; then
    nohup "./build/bin/db-desktop.app/Contents/MacOS/A modern database management tool for MySQL, Redis, and ClickHouse" > /dev/null 2>&1 &
    sleep 3
    if pgrep -f "db-desktop" > /dev/null; then
        echo "✅ 应用启动成功"
    else
        echo "❌ 应用启动失败"
        exit 1
    fi
else
    echo "❌ 编译后的应用文件不存在"
    exit 1
fi

# 3. 检查查询历史
echo -e "\n3. 检查查询历史状态..."
TOTAL_QUERIES=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "当前查询历史记录数: $TOTAL_QUERIES"

# 4. 清理测试数据
echo -e "\n4. 清理之前的测试数据..."
sqlite3 ~/.db-desktop/conversations.db "DELETE FROM query_history WHERE query = 'SELECT 1';"

# 5. 验证清理结果
TOTAL_AFTER_CLEANUP=$(sqlite3 ~/.db-desktop/conversations.db "SELECT COUNT(*) FROM query_history;")
echo "清理后查询历史记录数: $TOTAL_AFTER_CLEANUP"

echo -e "\n=== 修复验证完成 ==="
echo "现在请在应用中："
echo "1. 连接到MySQL数据库"
echo "2. 执行一个SQL查询"
echo "3. 检查查询历史页面是否显示记录"
