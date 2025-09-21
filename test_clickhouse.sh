#!/bin/bash

echo "测试ClickHouse连接..."

# 测试原生协议
echo "1. 测试原生协议 (9000端口):"
echo "SELECT 1" | nc -w 5 47.120.42.112 9000
echo ""

# 测试HTTP协议
echo "2. 测试HTTP协议 (8123端口):"
curl -s --max-time 5 "http://47.120.42.112:8123/?query=SELECT%201" || echo "HTTP连接失败"
echo ""

# 测试telnet连接
echo "3. 测试telnet连接:"
echo "quit" | telnet 47.120.42.112 9000 2>/dev/null | head -5
echo ""

echo "测试完成"
