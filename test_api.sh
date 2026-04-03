#!/bin/bash

echo "=== Guardian API 测试 ==="
echo ""

echo "1. 测试获取所有进程信息（格式化输出）"
echo "GET http://localhost:8080/api/processes"
curl -s http://localhost:8080/api/processes
echo ""
echo ""

echo "2. 测试获取单个进程信息"
echo "GET http://localhost:8080/api/process/web-server"
curl -s http://localhost:8080/api/process/web-server
echo ""
echo ""

echo "3. 测试健康检查"
echo "GET http://localhost:8080/api/health"
curl -s http://localhost:8080/api/health
echo ""
echo ""

echo "=== 测试完成 ==="
