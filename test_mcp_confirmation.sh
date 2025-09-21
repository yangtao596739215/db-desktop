#!/bin/bash

echo "🧪 测试MCP工具确认功能"
echo "========================"

# 启动应用
echo "🚀 启动应用..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 3

echo "✅ 应用已启动"
echo ""
echo "📋 测试步骤："
echo "1. 在AI助手界面中发送消息，要求执行数据库操作"
echo "2. 观察是否出现工具确认卡片"
echo "3. 点击确认或拒绝按钮"
echo "4. 观察工具执行结果"
echo ""
echo "💡 示例消息："
echo "- '查询所有用户表的数据'"
echo "- '执行Redis命令 KEYS *'"
echo "- '显示数据库中的所有表'"
echo ""
echo "🔍 检查点："
echo "- 工具确认卡片是否正确显示"
echo "- 确认/拒绝按钮是否正常工作"
echo "- 工具执行结果是否正确显示"
echo ""
echo "按任意键退出测试..."
read -n 1
