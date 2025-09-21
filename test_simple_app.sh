#!/bin/bash

echo "🧪 Testing simplified app version..."

# 启动应用
echo "🚀 Starting simplified application..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 3

echo "✅ Simplified application should now be running"
echo "📝 This version includes:"
echo "  1. Basic React app with Ant Design"
echo "  2. Simple menu navigation"
echo "  3. No complex stores or event listeners"
echo "  4. No Wails runtime dependencies"

echo ""
echo "🎯 Check if the application window appears with:"
echo "  - Left sidebar with menu items (MySQL, Redis, ClickHouse, AI助手, etc.)"
echo "  - Main content area showing 'MySQL 查询界面'"
echo "  - No white screen"

echo ""
echo "🔍 If this works, the issue is with the complex stores or event system"
echo "   If still white screen, the issue is more fundamental"
