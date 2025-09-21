#!/bin/bash

# 测试新的SendMessageStreamWithCompleteResponse方法

echo "🧪 测试SendMessageStreamWithCompleteResponse方法"
echo "================================================"

# 启动应用
echo "🚀 启动应用..."
./db-desktop-test &
APP_PID=$!

# 等待应用启动
sleep 3

echo "✅ 应用已启动 (PID: $APP_PID)"

# 测试1: 普通流式响应
echo ""
echo "📝 测试1: 普通流式响应"
echo "发送消息: '你好，请介绍一下你自己'"
echo "预期: 实时流式返回，完成后将完整响应保存到SQLite"

# 测试2: 包含MCP的流式响应
echo ""
echo "📝 测试2: 包含MCP的流式响应"
echo "发送消息: '请查询Redis数据库中的所有key'"
echo "预期: 实时流式返回，检测到MCP后等待完整响应再执行MCP逻辑，并保存到SQLite"

# 测试3: 包含MCP的流式响应（MySQL）
echo ""
echo "📝 测试3: 包含MCP的流式响应（MySQL）"
echo "发送消息: '请查询MySQL数据库中的用户表'"
echo "预期: 实时流式返回，检测到MCP后等待完整响应再执行MCP逻辑，并保存到SQLite"

# 测试4: 工具确认流程
echo ""
echo "📝 测试4: 工具确认流程"
echo "发送消息: '请执行一个数据库查询'"
echo "预期: 流式返回，显示工具确认卡片，用户确认后继续流式响应，并保存到SQLite"

echo ""
echo "🎯 测试完成！"
echo "请检查应用日志以验证新方法的功能："
echo "1. 流式响应是否正常工作"
echo "2. 完整响应是否正确保存到SQLite"
echo "3. MCP工具调用是否正确处理"
echo "4. 工具确认卡片是否正常显示"

# 清理
echo ""
echo "🧹 清理测试环境..."
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo "✅ 测试完成！"
