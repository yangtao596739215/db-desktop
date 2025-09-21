#!/bin/bash

# 脚本用于将 WithFields 调用替换为直接使用 Infof/Errorf/Warnf

# 替换 ai_service.go 中的 WithFields 调用
sed -i '' 's/utils\.WarnWithFields(map\[string\]interface{}{"dbType": dbType}, "No connected database found for MCP tool")/utils.Warnf("No connected database found for MCP tool: dbType=%s", dbType)/g' backend/app/ai_service.go

# 替换简单的单字段 WithFields 调用
sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"toolCallsCount": len(choice\.Message\.ToolCalls)}, "Checking for tool calls")/utils.Infof("Checking for tool calls: toolCallsCount=%d", len(choice.Message.ToolCalls))/g' backend/app/ai_service.go

sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"toolCallID": toolCall\.ID}, "Tool call confirmed via card")/utils.Infof("Tool call confirmed via card: toolCallID=%s", toolCall.ID)/g' backend/app/ai_service.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to execute confirmed tool call")/utils.Errorf("Failed to execute confirmed tool call: %v", err)/g' backend/app/ai_service.go

sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"toolCallID": toolCall\.ID}, "🔄 Executing Redis command")/utils.Infof("🔄 Executing Redis command: toolCallID=%s", toolCall.ID)/g' backend/app/ai_service.go

sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"toolCallID": toolCall\.ID}, "🔄 Executing MySQL query")/utils.Infof("🔄 Executing MySQL query: toolCallID=%s", toolCall.ID)/g' backend/app/ai_service.go

sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"toolCallID": toolCall\.ID}, "🔄 Executing ClickHouse query")/utils.Infof("🔄 Executing ClickHouse query: toolCallID=%s", toolCall.ID)/g' backend/app/ai_service.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"toolCallID": toolCall\.ID}, "Database manager not available for MCP tool")/utils.Errorf("Database manager not available for MCP tool: toolCallID=%s", toolCall.ID)/g' backend/app/ai_service.go

# 替换请求体相关的 WithFields 调用
sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"body": prettyJSON\.String()}, "📤 Streaming request body")/utils.Infof("📤 Streaming request body: %s", prettyJSON.String())/g' backend/app/ai_service.go

sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{"body": string(reqBody)}, "📤 Streaming request body (raw)")/utils.Infof("📤 Streaming request body (raw): %s", string(reqBody))/g' backend/app/ai_service.go

echo "WithFields 替换完成"
