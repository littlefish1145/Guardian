#!/bin/bash

echo "=== Guardian 改进测试 ==="
echo ""

echo "1. 改进的日志格式"
echo "   - 包含时间戳（精确到毫秒）"
echo "   - 包含日志级别（INFO/WARN/ERROR/DEBUG）"
echo "   - 包含进程名标识"
echo "   - 包含资源信息（Memory/Goroutines）"
echo ""

echo "2. 僵尸进程检测和清理"
echo "   - 自动检测僵尸进程状态"
echo "   - 使用 wait4 系统调用清理僵尸进程"
echo "   - 超过最大重启次数后自动放弃"
echo ""

echo "3. 状态变化检测"
echo "   - 只在进程状态变化时打印日志"
echo "   - 减少日志噪音，提高可读性"
echo ""

echo "4. Debug 模式"
echo "   - 配置文件支持 debug: true"
echo "   - 显示健康检查失败的详细原因"
echo "   - 显示僵尸进程清理信息"
echo ""

echo "5. 进程状态改进"
echo "   - 停止的进程标记为 'stopped' 而非 'failed'"
echo "   - 更准确的状态表示"
echo ""

echo "6. 预期日志输出示例："
echo "   [2026-04-03 15:04:05.000] [INFO] [SYSTEM] Guardian process manager starting..."
echo "   [2026-04-03 15:04:05.000] [INFO] [SYSTEM] Debug mode enabled"
echo "   [2026-04-03 15:04:05.000] [INFO] [web-server] Process started successfully, PID: 15, Memory: 10.50MB, Goroutines: 5"
echo "   [2026-04-03 15:04:10.000] [WARN] [web-server] Health check failed, PID: 15, Failures: 1/3, Reason: HTTP request failed: connection refused"
echo "   [2026-04-03 15:04:10.000] [INFO] [zombie-test] State: stopped, PID: 16, Restarts: 1, Memory: 7.85MB, Goroutines: 6"
echo ""

echo "7. 启动 Guardian 进行测试..."
./guardian -config guardian.yaml
