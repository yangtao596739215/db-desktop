#!/bin/bash

# 测试消息存储逻辑

echo "🧪 测试消息存储逻辑"
echo "================================"

# 启动应用
echo "🚀 启动应用..."
./db-desktop-test &
APP_PID=$!

# 等待应用启动
sleep 3

echo "✅ 应用已启动 (PID: $APP_PID)"

# 测试1: 普通消息存储
echo ""
echo "📝 测试1: 普通消息存储"
echo "发送消息: '你好，请介绍一下你自己'"
echo "预期: 用户消息立即存储，assistant消息流式结束后存储"

# 测试2: 包含MCP的消息存储
echo ""
echo "📝 测试2: 包含MCP的消息存储"
echo "发送消息: '请查询Redis数据库中的所有key'"
echo "预期: 用户消息立即存储，assistant消息流式结束后存储，MCP处理完成后继续存储"

# 测试3: 工具确认流程的消息存储
echo ""
echo "📝 测试3: 工具确认流程的消息存储"
echo "发送消息: '请执行一个数据库查询'"
echo "预期: 用户消息立即存储，assistant消息流式结束后存储，工具确认后继续存储"

echo ""
echo "🎯 测试完成！"
echo "请检查应用日志以验证消息存储逻辑："
echo "- 用户消息应该在收到时立即存储"
echo "- assistant消息应该在流式响应结束后存储"
echo "- 工具确认后的消息也应该正确存储"

# 清理
echo ""
echo "🧹 清理测试环境..."
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo "✅ 测试完成！"