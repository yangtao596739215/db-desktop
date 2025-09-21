#!/bin/bash

echo "🧪 Testing AI Config Persistence..."

# 检查配置文件是否存在
CONFIG_FILE="$HOME/.db-desktop/ai_config.json"

if [ -f "$CONFIG_FILE" ]; then
    echo "✅ AI config file exists: $CONFIG_FILE"
    echo "📄 Config file contents:"
    cat "$CONFIG_FILE" | jq . 2>/dev/null || cat "$CONFIG_FILE"
else
    echo "❌ AI config file not found: $CONFIG_FILE"
    echo "📁 Checking directory structure:"
    ls -la "$HOME/.db-desktop/" 2>/dev/null || echo "Directory does not exist"
fi

echo ""
echo "🔍 Checking database connections config for comparison:"
CONN_FILE="$HOME/.db-desktop/connections.json"
if [ -f "$CONN_FILE" ]; then
    echo "✅ Database connections file exists: $CONN_FILE"
    echo "📄 Connections file contents:"
    cat "$CONN_FILE" | jq . 2>/dev/null || cat "$CONN_FILE"
else
    echo "❌ Database connections file not found: $CONN_FILE"
fi
