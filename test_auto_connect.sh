#!/bin/bash

echo "🧪 测试自动连接功能"
echo "===================="

# 停止现有应用
echo "🛑 停止现有应用..."
pkill -f "db-desktop" 2>/dev/null || true
sleep 2

# 检查配置文件
echo "📋 检查配置文件..."
if [ ! -f ~/.db-desktop/connections.json ]; then
    echo "❌ 配置文件不存在: ~/.db-desktop/connections.json"
    echo "📝 请先创建配置文件："
    echo "   1. 复制示例文件: cp connections_example.json ~/.db-desktop/connections.json"
    echo "   2. 编辑配置文件: nano ~/.db-desktop/connections.json"
    echo "   3. 根据您的数据库环境修改配置"
    echo ""
    read -p "是否现在创建示例配置文件？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        mkdir -p ~/.db-desktop
        cp connections_example.json ~/.db-desktop/connections.json
        echo "✅ 已创建示例配置文件"
        echo "📝 请编辑配置文件后重新运行测试"
        exit 0
    else
        exit 1
    fi
else
    echo "✅ 找到配置文件: ~/.db-desktop/connections.json"
fi

# 构建应用
echo "🔨 构建应用..."
wails build

if [ $? -ne 0 ]; then
    echo "❌ 构建失败"
    exit 1
fi

echo "✅ 构建成功"

# 启动应用
echo "🚀 启动应用..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 3

echo ""
echo "📋 测试步骤："
echo "1. 查看应用启动日志，应该看到："
echo "   - 📋 Found X saved connections in config file"
echo "   - 🔄 Auto-connecting to saved database connections..."
echo "   - 🔌 Attempting to connect to [连接名称] ([类型])..."
echo "   - ✅/❌ 连接结果"
echo ""
echo "2. 在连接管理界面查看连接状态"
echo "3. 在AI助手界面测试工具确认功能"
echo ""
echo "💡 注意："
echo "- 如果数据库服务未运行，连接会失败，但这是正常的"
echo "- 配置文件中的连接会被自动尝试连接"
echo "- 查看详细说明: 数据库连接配置说明.md"
echo ""
echo "按任意键退出测试..."
read -n 1
