#!/bin/bash

# 替换 card.go 中剩余的 WithFields 调用

# 使用更通用的替换模式
sed -i '' 's/utils\.InfoWithFields(map\[string\]interface{}{\([^}]*\)}, "\([^"]*\)")/utils.Infof("\2: \1")/g' backend/app/card.go

sed -i '' 's/utils\.ErrorWithFields(map\[string\]interface{}{\([^}]*\)}, "\([^"]*\)")/utils.Errorf("\2: \1")/g' backend/app/card.go

sed -i '' 's/utils\.WarnWithFields(map\[string\]interface{}{\([^}]*\)}, "\([^"]*\)")/utils.Warnf("\2: \1")/g' backend/app/card.go

# 清理格式，将字段名和值分离
sed -i '' 's/"\([^"]*\)": \([^,}]*\),/\\1=\\2,/g' backend/app/card.go
sed -i '' 's/"\([^"]*\)": \([^,}]*\)}/\\1=\\2/g' backend/app/card.go

echo "card.go 剩余 WithFields 替换完成"
