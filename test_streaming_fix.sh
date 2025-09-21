#!/bin/bash

echo "🧪 Testing streaming response fix..."

# 构建应用
echo "📦 Building application..."
go build -o db-desktop-test .

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# 启动应用进行测试
echo "🚀 Starting application for testing..."
./db-desktop-test &

# 等待应用启动
sleep 3

# 获取应用PID
APP_PID=$!

echo "📱 Application started with PID: $APP_PID"

# 等待用户测试
echo "🔍 Please test the AI assistant with a query that should trigger tool calls (e.g., '查询用户表数据')"
echo "📊 Check the logs for streaming response processing details"
echo "⏹️  Press Ctrl+C to stop the test"

# 等待用户中断
trap "echo '🛑 Stopping test...'; kill $APP_PID 2>/dev/null; exit 0" INT

# 保持运行
wait $APP_PID
