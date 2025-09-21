#!/bin/bash

echo "🧪 Testing complete app with event listeners..."

# 启动应用
echo "🚀 Starting complete application..."
./build/bin/db-desktop.app/Contents/MacOS/A\ modern\ database\ management\ tool\ for\ MySQL,\ Redis,\ and\ ClickHouse &

# 等待应用启动
sleep 5

echo "✅ Complete application should now be running"
echo "📝 This version includes:"
echo "  1. All original components and stores"
echo "  2. Connection management"
echo "  3. AI Assistant store with event listeners"
echo "  4. All views (Query, Settings, AIAssistant, etc.)"
echo "  5. MsgVo-based event system"

echo ""
echo "🎯 Check if the application window appears with:"
echo "  - Left sidebar with all menu items"
echo "  - Main content area working"
echo "  - All tabs functional"
echo "  - No white screen"
echo "  - Console shows 'Event listeners initialized successfully'"

echo ""
echo "🔍 Test the AI Assistant feature:"
echo "  1. Click on 'AI助手' tab"
echo "  2. Try sending a message"
echo "  3. Check if streaming works with MsgVo events"

echo ""
echo "📊 Check browser console for:"
echo "  - 'Attempting to initialize event listeners...'"
echo "  - 'Event listeners initialized successfully'"
echo "  - Any error messages"
