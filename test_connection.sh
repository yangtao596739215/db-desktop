#!/bin/bash

echo "🔍 测试 ClickHouse 连接修复..."

# 测试 ClickHouse 是否在运行
echo "📡 检查 ClickHouse 服务状态..."
if nc -z localhost 9000; then
    echo "✅ ClickHouse 服务正在运行 (端口 9000)"
else
    echo "❌ ClickHouse 服务未运行"
    exit 1
fi

# 测试连接
echo "🔗 测试 ClickHouse 连接..."
if command -v clickhouse-client &> /dev/null; then
    echo "📊 ClickHouse 版本信息:"
    clickhouse-client --query "SELECT version()" 2>/dev/null || echo "❌ 无法连接到 ClickHouse"
else
    echo "⚠️  clickhouse-client 未安装，无法测试连接"
fi

echo "🏁 测试完成!"
echo ""
echo "💡 现在可以在应用中测试 ClickHouse 连接："
echo "   1. 打开应用"
echo "   2. 进入连接管理"
echo "   3. 选择 ClickHouse"
echo "   4. 添加新连接，使用以下配置："
echo "      - 主机: localhost"
echo "      - 端口: 9000"
echo "      - 用户: default"
echo "      - 密码: (留空)"
echo "      - 数据库: default"
