#!/bin/bash

# AI助手测试脚本
echo "=== AI助手功能测试 ==="

# 检查应用是否已构建
if [ ! -f "./db-desktop" ]; then
    echo "❌ 应用未构建，请先运行 'wails build'"
    exit 1
fi

echo "✅ 应用已构建"

# 测试AI配置
echo ""
echo "=== 测试AI配置 ==="
echo "请确保已配置以下环境变量："
echo "- OPENAI_API_KEY: 你的OpenAI API密钥"
echo "- OPENAI_BASE_URL: API基础URL (默认: https://api.openai.com/v1)"
echo ""

# 检查环境变量
if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  警告: OPENAI_API_KEY 环境变量未设置"
    echo "   请设置: export OPENAI_API_KEY='your-api-key'"
else
    echo "✅ OPENAI_API_KEY 已设置"
fi

if [ -z "$OPENAI_BASE_URL" ]; then
    echo "ℹ️  OPENAI_BASE_URL 未设置，将使用默认值"
    export OPENAI_BASE_URL="https://api.openai.com/v1"
else
    echo "✅ OPENAI_BASE_URL 已设置: $OPENAI_BASE_URL"
fi

echo ""
echo "=== 测试建议 ==="
echo "1. 启动应用: ./db-desktop"
echo "2. 在应用中添加数据库连接"
echo "3. 配置AI助手的API密钥和基础URL"
echo "4. 尝试以下测试用例："
echo ""
echo "   📝 测试用例1: 基础对话"
echo "   发送消息: '你好，请介绍一下你自己'"
echo ""
echo "   📝 测试用例2: Redis命令"
echo "   发送消息: '请帮我执行Redis命令 SET test_key hello_world'"
echo "   (需要先添加并连接Redis数据库)"
echo ""
echo "   📝 测试用例3: MySQL查询"
echo "   发送消息: '请帮我查询MySQL数据库中的用户表'"
echo "   (需要先添加并连接MySQL数据库)"
echo ""
echo "   📝 测试用例4: ClickHouse查询"
echo "   发送消息: '请帮我查询ClickHouse数据库中的日志表'"
echo "   (需要先添加并连接ClickHouse数据库)"
echo ""
echo "   📝 测试用例5: 智能查询建议"
echo "   发送消息: '我想查看最近一周的订单数据，应该怎么查询？'"
echo ""

echo "=== 功能特性 ==="
echo "✅ AI助手支持以下功能："
echo "   - 与OpenAI API集成"
echo "   - 支持流式和非流式响应"
echo "   - 自动调用数据库MCP工具"
echo "   - 支持Redis、MySQL、ClickHouse三种数据库"
echo "   - 自动查找已连接的数据库（无需指定连接ID）"
echo "   - 智能查询建议和自动执行"
echo "   - 友好的结果格式化显示"
echo ""

echo "=== 注意事项 ==="
echo "⚠️  重要提醒："
echo "   - 确保API密钥有效且有足够的配额"
echo "   - 数据库连接需要先配置并测试通过"
echo "   - 流式响应模式下不支持工具调用"
echo "   - 工具调用仅在非流式模式下可用"
echo ""

echo "测试完成！请按照上述建议进行实际测试。"
