#!/bin/bash

# 替换 app.go 中的 WithFields 调用

# 替换所有简单的 ErrorWithFields 调用
sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to get conversation messages")/utils.Errorf("Failed to get conversation messages: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to save user message")/utils.Errorf("Failed to save user message: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to save AI response")/utils.Errorf("Failed to save AI response: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to save tool result to database")/utils.Errorf("Failed to save tool result to database: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to get query history")/utils.Errorf("Failed to get query history: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to get query history by database type")/utils.Errorf("Failed to get query history by database type: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to get query history statistics")/utils.Errorf("Failed to get query history statistics: %v", err)/g' backend/app/app.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to clear query history")/utils.Errorf("Failed to clear query history: %v", err)/g' backend/app/app.go

echo "app.go WithFields 替换完成"
