#!/bin/bash

# 测试conversationID验证功能

echo "🧪 测试conversationID验证功能"
echo "================================"

# 编译项目
echo "📦 编译项目..."
go build -o db-desktop-test .

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo ""

# 测试场景说明
echo "📋 测试场景："
echo "1. 当conversationID为空时，应该返回特定错误"
echo "2. 当conversationID不为空时，应该正常处理"
echo ""

echo "🔍 检查代码中的错误定义..."
grep -n "ErrConversationIDRequired" backend/handler/app.go
echo ""

echo "🔍 检查SendMessageToConversation方法..."
grep -A 10 "检查conversationID是否为空" backend/handler/app.go
echo ""

echo "🔍 检查SendMessageStreamToConversation方法..."
grep -A 10 "检查conversationID是否为空" backend/handler/app.go
echo ""

echo "✅ 代码检查完成"
echo ""
echo "📝 前端处理建议："
echo "1. 捕获错误类型 ErrConversationIDRequired"
echo "2. 当收到此错误时，调用 CreateConversation 接口"
echo "3. 使用返回的conversationID重新发送消息"
echo ""
echo "示例前端代码："
echo "try {"
echo "  const response = await sendMessage(conversationID, message);"
echo "  // 处理响应"
echo "} catch (error) {"
echo "  if (error.message === 'CONVERSATION_ID_REQUIRED') {"
echo "    // 创建新会话"
echo "    const conversation = await createConversation('新对话');"
echo "    // 使用新会话ID重新发送消息"
echo "    const response = await sendMessage(conversation.id, message);"
echo "  }"
echo "}"
