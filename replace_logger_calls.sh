#!/bin/bash

# 批量替换ai_service.go中的logger调用

echo "🔄 批量替换ai_service.go中的logger调用..."

# 替换简单的Info调用
sed -i '' 's/s\.logger\.Info(/utils.Infof(/g' backend/app/ai_service.go
sed -i '' 's/s\.logger\.Debug(/utils.Debugf(/g' backend/app/ai_service.go
sed -i '' 's/s\.logger\.Warn(/utils.Warnf(/g' backend/app/ai_service.go
sed -i '' 's/s\.logger\.Error(/utils.Errorf(/g' backend/app/ai_service.go

# 替换WithField调用
sed -i '' 's/s\.logger\.WithField(/utils.WithField(/g' backend/app/ai_service.go

# 替换WithFields调用
sed -i '' 's/s\.logger\.WithFields(/utils.WithFields(/g' backend/app/ai_service.go

# 替换WithError调用
sed -i '' 's/s\.logger\.WithError(/utils.WithError(/g' backend/app/ai_service.go

echo "✅ 替换完成！"
