#!/bin/bash

# DB Desktop 启动脚本

echo "🚀 启动 DB Desktop..."

# 检查是否在开发模式
if [ "$1" = "dev" ]; then
    echo "📝 开发模式启动..."
    wails dev
else
    echo "🏗️ 构建并启动应用..."
    
    # 构建前端
    echo "📦 构建前端..."
    cd frontend
    npm run build
    cd ..
    
    # 构建应用
    echo "🔨 构建应用..."
    wails build
    
    # 启动应用
    echo "🎯 启动应用..."
    open "build/bin/db-desktop.app"
fi

echo "✅ 完成！"
