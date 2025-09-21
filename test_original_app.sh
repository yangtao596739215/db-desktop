#!/bin/bash

echo "🧪 Testing original app without event listeners..."

# 启动应用
echo "🚀 Starting original application..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 3

echo "✅ Original application should now be running"
echo "📝 This version includes:"
echo "  1. All original components and stores"
echo "  2. Connection management"
echo "  3. AI Assistant store (but no event listeners)"
echo "  4. All views (Query, Settings, AIAssistant, etc.)"

echo ""
echo "🎯 Check if the application window appears with:"
echo "  - Left sidebar with all menu items"
echo "  - Main content area working"
echo "  - All tabs functional (MySQL, Redis, ClickHouse, AI助手, etc.)"
echo "  - No white screen"

echo ""
echo "🔍 If this works, we can gradually enable event listeners"
echo "   If still white screen, the issue is with one of the stores or components"
