#!/bin/bash

# 替换 card.go 中的 WithFields 调用

# 替换复杂的 InfoWithFields 调用
sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{\n\t\t"cardID":         cardID,\n\t\t"showContent":    showContent,\n\t\t"conversationID": conversationID,\n\t\t"toolCallID":     toolCallID,\n\t\t"expiresAt":      card\.ExpiresAt,\n\t}, "Created confirmation card with metadata")/utils.Infof("Created confirmation card with metadata: cardID=%s, showContent=%s, conversationID=%s, toolCallID=%s, expiresAt=%s", cardID, showContent, conversationID, toolCallID, card.ExpiresAt)/g' backend/app/card.go

# 替换简单的 ErrorWithFields 调用
sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to create confirmation card")/utils.Errorf("Failed to create confirmation card: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to save confirmation card")/utils.Errorf("Failed to save confirmation card: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to load confirmation cards")/utils.Errorf("Failed to load confirmation cards: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to save confirmation cards")/utils.Errorf("Failed to save confirmation cards: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to delete confirmation card")/utils.Errorf("Failed to delete confirmation card: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to update confirmation card")/utils.Errorf("Failed to update confirmation card: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to get confirmation card")/utils.Errorf("Failed to get confirmation card: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to confirm tool call")/utils.Errorf("Failed to confirm tool call: %v", err)/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{"error": err}, "Failed to reject tool call")/utils.Errorf("Failed to reject tool call: %v", err)/g' backend/app/card.go

echo "card.go WithFields 替换完成"
