#!/bin/bash

echo "🧪 Testing safe event listener initialization..."

# 启动应用
echo "🚀 Starting application with safe event initialization..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 3

echo "✅ Application should now be running"
echo "📝 This version includes:"
echo "  1. App starts without event listeners (no white screen)"
echo "  2. Event listeners only initialize when AI Assistant tab is opened"
echo "  3. Safer initialization timing"

echo ""
echo "🎯 Test steps:"
echo "  1. Check if app starts normally (no white screen)"
echo "  2. Navigate through different tabs (MySQL, Redis, etc.)"
echo "  3. Click on 'AI助手' tab"
echo "  4. Check console for 'Initializing event listeners in AI Assistant component...'"
echo "  5. Try sending a message in AI Assistant"

echo ""
echo "🔍 Expected behavior:"
echo "  - App starts immediately without white screen"
echo "  - All tabs work normally"
echo "  - Event listeners only initialize when AI Assistant is accessed"
echo "  - AI Assistant streaming should work after initialization"
