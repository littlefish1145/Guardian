# 故障排查

常见问题和解决方案。

## 进程无法启动

### 检查配置文件语法

```bash
guardian -config guardian.yaml --validate
```

### 查看详细日志

```bash
./guardian -config guardian.yaml -log-level debug
```

### 检查进程依赖

```bash
curl http://localhost:8080/api/processes | jq '.[] | {name, state, failure_count}'
```

### 常见原因

1. **命令路径错误**: 确认命令在容器中可用
2. **权限不足**: 检查是否需要 root 权限
3. **依赖进程未启动**: 检查 `depends_on` 配置
4. **资源限制过严**: 调整 `resources` 配置

## 健康检查失败

### 手动测试健康检查端点

```bash
curl -v http://localhost:80/healthz
```

### 查看健康检查指标

```bash
curl http://localhost:9090/metrics | grep guardian_health_check
```

### 调整健康检查配置

```yaml
health_check:
  interval: 30s         # 增加检查间隔
  timeout: 10s          # 增加超时时间
  failure_threshold: 5  # 增加失败阈值
  initial_delay: 30s    # 增加首次检查延迟
```

### 常见原因

1. **端点未就绪**: 增加 `initial_delay`
2. **超时时间过短**: 增加 `timeout`
3. **端点路径错误**: 确认 URL 正确
4. **网络问题**: 检查端口和网络连通性

## 僵尸进程问题

### 查看僵尸进程重启次数

```bash
curl http://localhost:8080/api/processes | jq '.[] | {name, zombie_restarts}'
```

### 检查系统僵尸进程

```bash
ps aux | awk '$8 ~ /Z/ {print $0}'
```

### 调整僵尸进程配置

```yaml
restart_policy:
  zombie_check_enabled: true
  zombie_check_interval: 10s    # 缩短检测间隔
  zombie_max_restarts: 5        # 增加最大重启次数
```

## 资源限制不生效

### 确认 Cgroup v2 挂载

```bash
mount | grep cgroup
```

### 检查 Cgroup 配置

```bash
cat /sys/fs/cgroup/cgroup.controllers
```

### 查看 Guardian 日志中的资源控制信息

```bash
grep "resource" /var/log/guardian/*.log
```

### 常见原因

1. **Cgroup v2 未挂载**: 确保系统使用 Cgroup v2
2. **权限不足**: 需要 `SYS_ADMIN` 权限
3. **配置格式错误**: 检查 `memory_limit` 和 `cpu_quota` 格式

## 日志问题

### 日志文件过大

调整日志轮转配置：

```yaml
global:
  max_log_size_mb: 50
  max_log_backups: 3

processes:
  logging:
    max_size_mb: 50
    max_backups: 3
    compress: true
```

### 日志格式不正确

确认 JSON 格式配置：

```yaml
logging:
  json_format: true
  include_timestamp: true
```

### 查看实时日志

```bash
tail -f /var/log/guardian/*.log
```

## 性能问题

### 监控进程状态

```bash
watch -n 1 'curl -s http://localhost:9090/metrics | grep guardian_process_state'
```

### 持续监控进程状态

```bash
while true; do
  curl -s http://localhost:8080/api/processes | jq '.[] | {name, state, pid}'
  sleep 2
done
```

### 优化建议

1. **调整健康检查频率**: 避免过于频繁的检查
2. **优化重启策略**: 使用指数退避避免重启风暴
3. **合理设置资源限制**: 避免资源竞争

## 调试技巧

### 启用调试模式

```yaml
global:
  debug: true
  log_level: debug
```

### 查看实时日志

```bash
tail -f /var/log/guardian/*.log
```

### 监控指标

```bash
watch -n 1 'curl -s http://localhost:9090/metrics | grep guardian_process_state'
```

### 使用 API 调试

```bash
# 持续监控进程状态
while true; do
  curl -s http://localhost:8080/api/processes | jq '.[] | {name, state, pid}'
  sleep 2
done
```

## 容器环境问题

### Docker 容器无法启动

1. 检查端口冲突：
   ```bash
   docker ps -a | grep guardian
   ```

2. 查看容器日志：
   ```bash
   docker logs <container-id>
   ```

3. 检查卷挂载：
   ```bash
   docker inspect <container-id> | grep Mounts
   ```

### K8s Pod 无法启动

1. 查看 Pod 状态：
   ```bash
   kubectl describe pod <pod-name>
   ```

2. 查看容器日志：
   ```bash
   kubectl logs <pod-name> -c guardian
   ```

3. 检查权限配置：
   ```bash
   kubectl get pod <pod-name> -o yaml | grep securityContext
   ```

## 获取帮助

如果以上方法无法解决问题：

1. 查看 [GitHub Issues](https://github.com/littlefish1145/guardian/issues)
2. 发起新的 Issue 并提供详细信息
3. 查看日志和指标输出

## 相关文档

- [[配置指南 | Configuration]]
- [[API 文档 | API-Reference]]
