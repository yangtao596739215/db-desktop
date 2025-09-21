#!/bin/bash

echo "🔍 调试MCP工具确认功能"
echo "========================"

# 检查应用是否在运行
if pgrep -f "db-desktop" > /dev/null; then
    echo "✅ 应用正在运行"
else
    echo "❌ 应用未运行，请先启动应用"
    exit 1
fi

echo ""
echo "📋 测试步骤："
echo "1. 确保已配置AI API Key"
echo "2. 确保已连接数据库（MySQL/Redis/ClickHouse）"
echo "3. 在AI助手界面发送以下测试消息："
echo ""
echo "测试消息示例："
echo "- '查询所有用户表的数据'"
echo "- '执行Redis命令 KEYS *'"
echo "- '显示数据库中的所有表'"
echo "- '帮我查询用户表的结构'"
echo ""
echo "🔍 调试信息："
echo "- 打开浏览器开发者工具（F12）"
echo "- 查看Console标签页的日志输出"
echo "- 查找 'Pending tool calls:' 和 'pendingToolCalls updated:' 日志"
echo ""
echo "如果看到工具调用日志但没看到卡片，可能是以下原因："
echo "1. AI没有返回工具调用（需要数据库连接）"
echo "2. 工具调用格式不正确"
echo "3. 前端渲染问题"
echo ""
echo "按任意键继续..."
read -n 1
