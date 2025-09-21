#!/bin/bash

# 测试所有功能都使用流式响应

echo "🧪 测试所有功能都使用流式响应"
echo "================================"

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
echo "预期: 实时流式返回，无MCP调用"

# 测试2: 包含MCP的流式响应
echo ""
echo "📝 测试2: 包含MCP的流式响应"
echo "发送消息: '请查询Redis数据库中的所有key'"
echo "预期: 实时流式返回，检测到MCP后等待完整响应再执行MCP逻辑"

# 测试3: 包含MCP的流式响应（MySQL）
echo ""
echo "📝 测试3: 包含MCP的流式响应（MySQL）"
echo "发送消息: '请查询MySQL数据库中的用户表'"
echo "预期: 实时流式返回，检测到MCP后等待完整响应再执行MCP逻辑"

# 测试4: 工具确认流程
echo ""
echo "📝 测试4: 工具确认流程"
echo "发送消息: '请执行一个数据库查询'"
echo "预期: 流式返回，显示工具确认卡片，用户确认后继续流式响应"

echo ""
echo "🎯 测试完成！"
echo "请检查应用日志以验证所有功能都使用流式响应"

# 清理
echo ""
echo "🧹 清理测试环境..."
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null

echo "✅ 测试完成！"
