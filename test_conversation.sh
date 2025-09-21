#!/bin/bash

echo "🧪 Testing conversation management functionality..."

# 测试创建会话
echo "📝 Testing conversation creation..."
curl -X POST http://localhost:8080/api/conversations \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Conversation"}' \
  || echo "❌ Conversation creation test failed"

# 测试获取会话列表
echo "📋 Testing conversation listing..."
curl -X GET http://localhost:8080/api/conversations \
  || echo "❌ Conversation listing test failed"

echo "✅ Conversation management tests completed!"
