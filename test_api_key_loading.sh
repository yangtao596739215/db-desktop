#!/bin/bash

echo "🔍 测试API Key加载"
echo "================================"

# 检查配置文件
echo "1. 检查配置文件内容:"
cat ~/.db-desktop/ai_config.json
echo ""

# 检查应用日志
echo "2. 检查应用启动日志:"
# 这里我们需要通过前端界面来触发AI请求，或者直接查看应用日志

echo "3. 通过前端界面测试:"
echo "   - 打开AI助手界面"
echo "   - 发送消息: '查询user表信息'"
echo "   - 观察日志中的Authorization header"

echo ""
echo "🔧 预期结果:"
echo "================================"
echo "1. 配置文件应该包含正确的API Key"
echo "2. 应用启动时应该显示 'Loaded AI config from file' 日志"
echo "3. 请求头应该显示正确的API Key（不是test-key）"

echo ""
echo "📋 如果仍然显示test-key，可能的原因:"
echo "================================"
echo "1. 配置文件格式问题"
echo "2. 配置加载时机问题"
echo "3. 配置被覆盖"

echo ""
echo "✨ 请通过前端界面测试，然后查看日志输出"
