# 配置指南

详细的 Guardian 配置说明。

## 全局配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `log_level` | string | `info` | 日志级别 (debug, info, warn, error) |
| `metrics_port` | int | `9090` | Prometheus 指标服务端口 |
| `api_port` | int | `8080` | RESTful API 服务端口 |
| `log_dir` | string | `/var/log/guardian` | 日志目录 |
| `max_log_size_mb` | int | `100` | 单个日志文件最大 MB |
| `max_log_backups` | int | `5` | 保留的日志备份数量 |
| `debug` | bool | `false` | 调试模式，输出详细日志 |
| `shutdown_timeout` | duration | `30s` | 优雅关闭超时时间 |

## 进程配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 进程名称（唯一标识） |
| `command` | []string | 是 | 要执行的命令及参数 |
| `working_dir` | string | 否 | 工作目录 |
| `depends_on` | []string | 否 | 依赖的进程列表（按顺序启动） |
| `environment` | map[string]string | 否 | 环境变量 |
| `user` | string | 否 | 运行用户（需要 root 权限） |

## 健康检查配置

### 基本配置

```yaml
health_check:
  type: http  # http, tcp, exec
  endpoint: http://localhost:80/healthz
  interval: 10s
  timeout: 3s
  failure_threshold: 3
  success_threshold: 1
  initial_delay: 5s
```

### 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `type` | string | - | 检查类型：`http`、`tcp`、`exec` |
| `endpoint` | string | - | 检查端点（HTTP URL 或 TCP 地址） |
| `interval` | duration | `10s` | 检查间隔 |
| `timeout` | duration | `3s` | 超时时间 |
| `failure_threshold` | int | `3` | 连续失败次数阈值（-1 表示无限重试） |
| `success_threshold` | int | `1` | 连续成功恢复次数 |
| `initial_delay` | duration | `0s` | 首次检查延迟 |

### 检查类型示例

**HTTP 检查**:
```yaml
health_check:
  type: http
  endpoint: http://localhost:8080/health
  interval: 10s
  timeout: 3s
```

**TCP 检查**:
```yaml
health_check:
  type: tcp
  endpoint: localhost:9000
  interval: 10s
  timeout: 3s
```

**Exec 检查**:
```yaml
health_check:
  type: exec
  command: ["pgrep", "-f", "myapp"]
  interval: 30s
  timeout: 5s
```

## 重启策略

```yaml
restart_policy:
  policy: on-failure  # always, on-failure, never
  max_restarts: 5
  base_delay: 1s
  max_delay: 60s
  zombie_max_restarts: 3
  zombie_check_enabled: true
  zombie_check_interval: 30s
```

### 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `policy` | string | `on-failure` | 重启策略：`always`、`on-failure`、`never` |
| `max_restarts` | int | `5` | 最大重启次数 |
| `base_delay` | duration | `1s` | 初始重启延迟（支持指数退避） |
| `max_delay` | duration | `60s` | 最大重启延迟 |
| `zombie_max_restarts` | int | `3` | 僵尸进程最大重启次数 |
| `zombie_check_enabled` | bool | `true` | 启用僵尸进程检测 |
| `zombie_check_interval` | duration | `30s` | 僵尸进程检测间隔 |

### 重启策略说明

- **always**: 总是重启，无论进程是否正常退出
- **on-failure**: 仅在进程失败时重启
- **never**: 从不重启

### 指数退避

实际重启延迟计算公式：
```
实际延迟 = min(base_delay * 2^(restart_count-1), max_delay)
```

示例：
- 第 1 次重启：1s
- 第 2 次重启：2s
- 第 3 次重启：4s
- 第 4 次重启：8s
- ...
- 最大不超过 max_delay

## 资源限制（Cgroup v2）

```yaml
resources:
  memory_limit: 512MB  # 或 1GB
  cpu_quota: "100%"    # 或 200%
  nofile: 65535
```

### 参数说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `memory_limit` | string | 内存限制（支持 MB/GB 单位） |
| `cpu_quota` | string | CPU 配额（百分比或核心数） |
| `nofile` | int | 最大打开文件描述符数 |

## 日志配置

```yaml
logging:
  stdout: true
  stderr: true
  file: /var/log/guardian/web.log
  max_size_mb: 100
  max_backups: 5
  compress: true
  compress_delay: 24h
  json_format: true
  include_timestamp: true
```

### 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `stdout` | bool | `true` | 捕获 stdout |
| `stderr` | bool | `true` | 捕获 stderr |
| `file` | string | - | 日志文件路径 |
| `max_size_mb` | int | `100` | 单文件最大 MB |
| `max_backups` | int | `5` | 保留备份数 |
| `compress` | bool | `true` | 轮转时压缩 |
| `compress_delay` | duration | `24h` | 压缩延迟 |
| `json_format` | bool | `true` | JSON 格式 |
| `include_timestamp` | bool | `true` | 包含时间戳 |

## 配置示例

### 生产环境配置

```yaml
global:
  log_level: warn
  metrics_port: 9090
  api_port: 8080
  log_dir: /var/log/guardian
  max_log_size_mb: 500
  max_log_backups: 10
  debug: false

processes:
  - name: nginx
    command: ["/usr/sbin/nginx", "-g", "daemon off;"]
    
    health_check:
      type: http
      endpoint: http://localhost:80/healthz
      interval: 5s
      timeout: 2s
      failure_threshold: 2
    
    restart_policy:
      policy: on-failure
      max_restarts: 10
      base_delay: 5s
      max_delay: 120s
    
    resources:
      memory_limit: 1GB
      cpu_quota: "200%"
      nofile: 65535
```

### 开发环境配置

```yaml
global:
  log_level: debug
  metrics_port: 9090
  api_port: 8080
  debug: true

processes:
  - name: app
    command: ["./myapp", "--dev"]
    
    health_check:
      type: http
      endpoint: http://localhost:8080/health
      interval: 30s
      timeout: 5s
      failure_threshold: 5
    
    restart_policy:
      policy: always
      max_restarts: 100
      base_delay: 1s
```

### 多进程管理配置

```yaml
processes:
  # 主应用
  - name: app
    command: ["/app/server"]
    health_check:
      type: http
      endpoint: http://localhost:8080/health
  
  # 日志收集 sidecar
  - name: log-collector
    command: ["/usr/bin/fluent-bit", "-c", "/fluent-bit.conf"]
    depends_on:
      - app
  
  # 监控 agent
  - name: metrics-agent
    command: ["/usr/bin/datadog-agent"]
    depends_on:
      - app
```

## 相关文档

- [[快速开始 | Getting-Started]]
- [[API 文档 | API-Reference]]
- [[故障排查 | Troubleshooting]]
