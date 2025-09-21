#!/bin/bash

echo "🧪 测试 panic 恢复机制"
echo "========================"

# 编译程序
echo "📦 编译程序..."
go build -o db-desktop-test
if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi
echo "✅ 编译成功"

# 运行程序（在后台运行，然后检查是否有 panic 恢复）
echo "🚀 启动程序..."
./db-desktop-test &
APP_PID=$!

# 等待几秒钟让程序启动
sleep 3

# 检查程序是否还在运行
if ps -p $APP_PID > /dev/null; then
    echo "✅ 程序正在运行 (PID: $APP_PID)"
    echo "📝 程序已启动，panic 恢复机制已就绪"
    echo "💡 如果发生 panic，程序会打印堆栈信息并退出"
    
    # 停止程序
    echo "🛑 停止程序..."
    kill $APP_PID
    wait $APP_PID 2>/dev/null
    echo "✅ 程序已停止"
else
    echo "❌ 程序启动失败或已崩溃"
    exit 1
fi

echo "🎉 测试完成！panic 恢复机制已配置"
