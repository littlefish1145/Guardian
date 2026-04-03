# Guardian

<div align="center">

**云原生轻量级进程守护程序，专为容器环境设计**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-EA8139?style=flat&logo=prometheus)](https://prometheus.io/)

[快速开始](#快速开始) • [功能特性](#功能特性) • [配置指南](#配置指南) • [API 文档](#api-文档) • [使用场景](#使用场景)

</div>

---

## 📖 简介

Guardian 是 `systemd` 和 `supervisord` 的现代替代方案，专为容器场景打造。它作为 PID 1 运行在容器内，负责管理多个进程、处理健康检查、限制资源使用、导出 Prometheus 指标。

**核心特点**：单静态二进制文件（约 15MB）、零依赖、即插即用。

### 🎯 设计目标

- **容器原生**：为容器环境而生，完美适配 Docker 和 Kubernetes
- **轻量级**：单静态二进制，无外部依赖，启动迅速
- **可观测性**：内置 Prometheus 指标，实时监控进程状态
- **高可用**：自动健康检查和重启策略，确保服务持续运行
- **资源隔离**：基于 Cgroup v2 的精细资源控制

## ✨ 功能特性

| 功能 | 描述 |
|------|------|
| **进程管理** | 启动、停止、重启进程，支持依赖顺序 |
| **健康检查** | HTTP/TCP/Exec 健康检查，可配置阈值 |
| **自动重启** | 可配置重启策略（always / on-failure / never） |
| **日志管理** | JSON 结构化日志，支持自动轮转 |
| **Prometheus 指标** | 内置 metrics 服务器 |
| **资源限制** | 基于 Cgroup v2 的 CPU/内存/文件描述符限制 |
| **优雅关闭** | 信号处理，支持进程优雅终止 |
| **零依赖** | 单静态二进制，无需 shell |
| **僵尸进程处理** | 自动检测和回收僵尸进程 |
| **API 接口** | RESTful API 查询进程状态 |

### 🔍 核心功能详解

#### 1. 进程生命周期管理
- 支持多进程并发启动，按依赖顺序编排
- 进程状态机：stopped → starting → running → stopping
- 自动重启策略：always / on-failure / never
- 指数退避重启延迟，避免重启风暴

#### 2. 健康检查系统
- **HTTP 检查**：支持 GET/POST 请求，验证 HTTP 状态码
- **TCP 检查**：端口连通性检测
- **Exec 检查**：执行自定义命令判断健康状态
- 可配置失败阈值，支持无限重试模式（failure_threshold: -1）

#### 3. 僵尸进程自动回收
- 定期检测进程是否为僵尸状态
- 自动重启僵尸进程，保持服务可用
- 可配置最大重启次数，避免无限重启循环
- 独立的僵尸进程计数器，便于监控

#### 4. 日志管理
- JSON 结构化日志，便于日志收集系统解析
- 自动日志轮转，防止日志文件过大
- 支持日志压缩，节省存储空间
- 分别捕获 stdout 和 stderr

#### 5. 资源隔离（Cgroup v2）
- CPU 配额限制（百分比或核心数）
- 内存限制（支持 MB/GB 单位）
- 文件描述符数量限制
- 避免单个进程占用过多系统资源

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         Guardian                             │
├─────────────────────────────────────────────────────────────┤
│  cmd/guardian/main.go                                        │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │
│  │  Config  │  │  Signal  │  │ Metrics  │  │    Log      │ │
│  │  Loader  │  │  Router  │  │ Server   │  │   Engine    │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────┬──────┘ │
│       │             │             │               │         │
│       └──────────────┴──────┬─────┴───────────────┘         │
│                             │                                │
│                    ┌────────▼────────┐                      │
│                    │ Process Manager  │                      │
│                    └────────┬────────┘                      │
│                             │                                │
│       ┌─────────────────────┼─────────────────────┐         │
│       │          ┌──────────┴──────────┐          │         │
│   ┌───▼───┐   ┌───▼───┐   ┌───▼───┐   ┌───▼───┐        │
│   │Proc 1 │   │Proc 2 │   │Proc 3 │   │Proc N │        │
│   └───────┘   └───────┘   └───────┘   └───────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- **Go** 1.21+
- **Docker** 20.10+（可选）
- **Make**（可选，方便构建）
- **Linux** 环境（推荐，完整支持 Cgroup v2）

### 编译

```bash
# 克隆仓库
git clone https://github.com/your-org/guardian.git
cd guardian

# 安装依赖
make install-deps

# 本地编译
make build

# Docker 编译
make docker-build

# 交叉编译（其他平台）
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o guardian
```

### 运行

#### 本地运行

```bash
# 使用默认配置运行
./guardian -config guardian.yaml

# 指定日志级别
./guardian -config guardian.yaml -log-level debug
```

#### Docker 运行

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run

# 或使用 docker 命令
docker run --rm -it \
  -p 9090:9090 \
  -p 8080:8080 \
  -v $(pwd)/guardian.yaml:/etc/guardian/guardian.yaml \
  guardian:latest
```

#### Kubernetes 部署

```bash
# 创建 ConfigMap
kubectl create configmap guardian-config --from-file=guardian.yaml

# 部署 Pod
kubectl apply -f k8s/deployment.yaml
```

### 配置文件示例

编辑 `guardian.yaml`:

```yaml
global:
  log_level: info
  metrics_port: 9090
  log_dir: /var/log/guardian
  max_log_size_mb: 100
  max_log_backups: 5

processes:
  - name: web-server
    command: ["/usr/sbin/nginx", "-g", "daemon off;"]
    working_dir: /usr/share/nginx/html

    health_check:
      type: http
      endpoint: http://localhost:80/healthz
      interval: 10s
      timeout: 3s
      failure_threshold: 3

    restart_policy:
      policy: on-failure
      max_restarts: 5

    resources:
      memory_limit: 512MB
      cpu_quota: "100%"
      nofile: 65535

    logging:
      stdout: true
      stderr: true
      file: /var/log/guardian/web.log
      max_size_mb: 100
      max_backups: 5
      compress: true

  - name: sidecar
    command: ["/usr/local/bin/sidecar"]
    depends_on:
      - web-server
```

## 📝 配置详解

### 全局配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `log_level` | string | `info` | 日志级别 (debug, info, warn, error) |
| `metrics_port` | int | `9090` | Prometheus 指标服务端口 |
| `api_port` | int | `8080` | RESTful API 服务端口 |
| `log_dir` | string | `/var/log/guardian` | 日志目录 |
| `max_log_size_mb` | int | `100` | 单个日志文件最大 MB |
| `max_log_backups` | int | `5` | 保留的日志备份数量 |
| `debug` | bool | `false` | 调试模式，输出详细日志 |

### 高级配置选项

### 高级配置选项

#### 僵尸进程处理配置

```yaml
restart_policy:
  policy: on-failure
  max_restarts: 5              # 普通重启最大次数
  zombie_max_restarts: 3       # 僵尸进程最大重启次数
  zombie_check_enabled: true   # 启用僵尸进程检测
  zombie_check_interval: 30s   # 僵尸进程检测间隔（默认 30 秒）
```

#### 指数退避配置

```yaml
restart_policy:
  policy: on-failure
  base_delay: 1s    # 初始重启延迟
  max_delay: 60s    # 最大重启延迟
  # 实际延迟 = min(base_delay * 2^(restart_count-1), max_delay)
```

#### 进程配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 进程名称（唯一标识） |
| `command` | []string | 是 | 要执行的命令及参数 |
| `working_dir` | string | 否 | 工作目录 |
| `depends_on` | []string | 否 | 依赖的进程列表（按顺序启动） |
| `environment` | map[string]string | 否 | 环境变量 |
| `user` | string | 否 | 运行用户（需要 root 权限） |

### 健康检查高级配置

```yaml
health_check:
  type: http
  endpoint: http://localhost:80/healthz
  interval: 10s
  timeout: 3s
  failure_threshold: 3      # 连续失败 3 次后重启
  success_threshold: 1      # 连续成功 1 次后恢复健康（默认 1）
  initial_delay: 5s         # 首次检查延迟（可选）
```

### 重启策略

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `policy` | string | `on-failure` | `always`、`on-failure`、`never` |
| `max_restarts` | int | `5` | 最大重启次数 |
| `base_delay` | duration | `1s` | 初始重启延迟（支持指数退避） |
| `max_delay` | duration | `60s` | 最大重启延迟 |
| `zombie_max_restarts` | int | `3` | 僵尸进程最大重启次数 |
| `zombie_check_enabled` | bool | `true` | 启用僵尸进程检测 |

### 资源限制

| 参数 | 类型 | 说明 |
|------|------|------|
| `memory_limit` | string | 内存限制（如 `512MB`、`1GB`） |
| `cpu_quota` | string | CPU 配额（如 `100%`、`200%`） |
| `nofile` | int | 最大打开文件描述符数 |

### 日志配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `stdout` | bool | `true` | 捕获 stdout |
| `stderr` | bool | `true` | 捕获 stderr |
| `file` | string | - | 日志文件路径 |
| `max_size_mb` | int | `100` | 单文件最大 MB |
| `max_backups` | int | `5` | 保留备份数 |
| `compress` | bool | `true` | 轮转时压缩 |

## 📊 API 文档

Guardian 提供 RESTful API 用于查询进程状态和管理。

### API 端点

#### 1. 获取所有进程信息

**端点:** `GET /api/processes`

**描述:** 获取所有进程的详细信息（格式化 JSON 输出）

**响应示例:**
```json
[
  {
    "name": "web-server",
    "state": "running",
    "pid": 15,
    "restarts": 0,
    "memory_mb": 10.50,
    "last_start": "2026-04-03T15:04:05Z",
    "failure_count": 0,
    "healthy_count": 10,
    "zombie_restarts": 0,
    "abandoned": false
  }
]
```

#### 2. 获取单个进程信息

**端点:** `GET /api/process/{name}`

**描述:** 根据进程名获取单个进程的详细信息

**参数:**
- `name` - 进程名称

**响应示例:**
```json
{
  "name": "web-server",
  "state": "running",
  "pid": 15,
  "restarts": 0,
  "memory_mb": 10.50,
  "last_start": "2026-04-03T15:04:05Z",
  "failure_count": 0,
  "healthy_count": 10,
  "zombie_restarts": 0,
  "abandoned": false
}
```

#### 3. 健康检查

**端点:** `GET /api/health`

**描述:** API 服务健康检查

**响应示例:**
```json
{
  "status": "ok",
  "time": "2026-04-03T15:04:05Z"
}
```

### 使用示例

```bash
# 查询所有进程
curl http://localhost:8080/api/processes

# 查询单个进程
curl http://localhost:8080/api/process/web-server

# 健康检查
curl http://localhost:8080/api/health

# 使用 jq 格式化输出
curl -s http://localhost:8080/api/processes | jq .
```

### 进程状态说明

| 状态 | 说明 |
|------|------|
| `starting` | 进程启动中 |
| `running` | 进程运行中 |
| `stopping` | 进程停止中 |
| `stopped` | 进程已停止 |
| `failed` | 进程失败 |
| `reclaimed` | 进程已回收（僵尸进程专用） |

#### 日志高级配置

```yaml
logging:
  stdout: true
  stderr: true
  file: /var/log/guardian/web.log
  max_size_mb: 100
  max_backups: 5
  compress: true           # 轮转时压缩
  compress_delay: 24h      # 压缩延迟（可选）
  json_format: true        # JSON 格式（默认 true）
  include_timestamp: true  # 包含时间戳
```

### 最佳实践配置示例

#### 生产环境配置

```yaml
global:
  log_level: warn          # 生产环境降低日志级别
  metrics_port: 9090
  api_port: 8080
  log_dir: /var/log/guardian
  max_log_size_mb: 500     # 生产环境增大日志文件
  max_log_backups: 10      # 保留更多备份
  debug: false

processes:
  - name: nginx
    command: ["/usr/sbin/nginx", "-g", "daemon off;"]
    
    # 健康检查：严格模式
    health_check:
      type: http
      endpoint: http://localhost:80/healthz
      interval: 5s         # 更频繁的检查
      timeout: 2s
      failure_threshold: 2 # 更快失败检测
    
    # 重启策略：保守模式
    restart_policy:
      policy: on-failure
      max_restarts: 10
      base_delay: 5s       # 更长的初始延迟
      max_delay: 120s
    
    # 资源限制
    resources:
      memory_limit: 1GB
      cpu_quota: "200%"
      nofile: 65535
```

#### 开发环境配置

```yaml
global:
  log_level: debug         # 开发环境详细日志
  metrics_port: 9090
  api_port: 8080
  debug: true

processes:
  - name: app
    command: ["./myapp", "--dev"]
    
    # 健康检查：宽松模式
    health_check:
      type: http
      endpoint: http://localhost:8080/health
      interval: 30s        # 较低频率
      timeout: 5s
      failure_threshold: 5
    
    # 重启策略：激进模式
    restart_policy:
      policy: always       # 总是重启
      max_restarts: 100
      base_delay: 1s
```

### 健康检查高级配置

## 📊 Prometheus 指标

Guardian 在 `http://localhost:9090/metrics` 暴露指标：

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `guardian_process_state` | Gauge | process | 进程状态 (0=停止，1=运行，2=失败) |
| `guardian_process_restarts_total` | Counter | process | 进程重启总次数 |
| `guardian_process_healthy` | Gauge | process | 健康检查状态 (0=不健康，1=健康) |
| `guardian_process_uptime_seconds` | Gauge | process | 进程运行时间（秒） |
| `guardian_health_check_failures_total` | Counter | process | 健康检查失败总次数 |
| `guardian_zombie_restarts_total` | Counter | process | 僵尸进程重启总次数 |
| `guardian_process_memory_bytes` | Gauge | process | 进程内存使用（字节） |
| `guardian_process_cpu_seconds_total` | Counter | process | 进程 CPU 使用总时间 |

### Grafana 仪表盘示例

```json
{
  "dashboard": {
    "title": "Guardian Process Monitor",
    "panels": [
      {
        "title": "Process State",
        "targets": [
          {
            "expr": "guardian_process_state",
            "legendFormat": "{{process}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "targets": [
          {
            "expr": "guardian_process_memory_bytes / 1024 / 1024",
            "legendFormat": "{{process}} MB"
          }
        ]
      }
    ]
  }
}
```

### 其他端点

- `/ready` - 就绪探针
- `/live` - 存活探针
- `/api/health` - API 健康检查

## 🐳 Docker 使用

Guardian 专为容器 PID 1 场景设计：

### Dockerfile 示例

```dockerfile
FROM nginx:alpine

# 复制 Guardian 二进制文件
COPY --from=builder /build/guardian /usr/local/bin/
COPY guardian.yaml /etc/guardian/guardian.yaml

# 暴露 metrics 和 API 端口
EXPOSE 9090 8080

# 设置入口点
ENTRYPOINT ["/usr/local/bin/guardian", "-config", "/etc/guardian/guardian.yaml"]
```

### 多阶段构建示例

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o guardian

# 运行阶段
FROM nginx:alpine

# 安装必要的依赖（如需要）
RUN apk add --no-cache python3

# 复制 Guardian
COPY --from=builder /build/guardian /usr/local/bin/
COPY guardian.yaml /etc/guardian/guardian.yaml

EXPOSE 9090 8080

ENTRYPOINT ["/usr/local/bin/guardian", "-config", "/etc/guardian/guardian.yaml"]
```

### Docker Compose 示例

```yaml
version: '3.8'

services:
  guardian:
    build: .
    ports:
      - "9090:9090"  # Metrics
      - "8080:8080"  # API
    volumes:
      - ./guardian.yaml:/etc/guardian/guardian.yaml:ro
      - ./logs:/var/log/guardian
    environment:
      - LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 运行

```bash
docker run --rm -it \
  -p 9090:9090 \
  -p 8080:8080 \
  -v $(pwd)/guardian.yaml:/etc/guardian/guardian.yaml \
  guardian:latest
```

## ☸️ Kubernetes 使用

Guardian 可作为 init 容器或 sidecar：

### 作为 Init 容器

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-with-guardian
  labels:
    app: myapp
spec:
  initContainers:
    - name: guardian
      image: guardian:latest
      securityContext:
        capabilities:
          add: ["SYS_ADMIN"]  # 需要 Cgroup 权限
      volumeMounts:
        - name: config
          mountPath: /etc/guardian
      ports:
        - containerPort: 9090
          name: metrics
        - containerPort: 8080
          name: api
  containers:
    - name: app
      image: myapp:latest
      # Guardian 会管理这个容器
  
  volumes:
    - name: config
      configMap:
        name: guardian-config
```

### 作为 Sidecar 容器

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-with-sidecar
spec:
  containers:
    - name: app
      image: myapp:latest
      lifecycle:
        preStop:
          exec:
            command: ["sleep", "30"]  # 优雅关闭
    
    - name: guardian
      image: guardian:latest
      securityContext:
        capabilities:
          add: ["SYS_ADMIN"]
      volumeMounts:
        - name: config
          mountPath: /etc/guardian
      ports:
        - containerPort: 9090
          name: metrics
        - containerPort: 8080
          name: api
  
  volumes:
    - name: config
      configMap:
        name: guardian-config
```

### Deployment 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guardian-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: guardian
  template:
    metadata:
      labels:
        app: guardian
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      containers:
        - name: guardian
          image: guardian:latest
          securityContext:
            capabilities:
              add: ["SYS_ADMIN"]
          volumeMounts:
            - name: config
              mountPath: /etc/guardian
          ports:
            - containerPort: 9090
              name: metrics
            - containerPort: 8080
              name: api
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /api/health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /ready
              port: 9090
            initialDelaySeconds: 5
            periodSeconds: 10
      
      volumes:
        - name: config
          configMap:
            name: guardian-config
```

### ServiceMonitor（Prometheus Operator）

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: guardian
  labels:
    app: guardian
spec:
  selector:
    matchLabels:
      app: guardian
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

## 🎯 使用场景

### 1. 微服务容器管理

在微服务架构中，一个容器可能需要运行多个相关进程：

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

### 2. 传统应用容器化

将需要多个进程的传统应用迁移到容器：

```yaml
processes:
  # Web 服务器
  - name: nginx
    command: ["/usr/sbin/nginx", "-g", "daemon off;"]
  
  # PHP-FPM
  - name: php-fpm
    command: ["/usr/sbin/php-fpm", "--nodaemonize"]
    depends_on:
      - nginx
  
  # 定时任务
  - name: cron
    command: ["/usr/sbin/cron", "-f"]
```

### 3. 数据处理管道

管理数据处理流程中的多个组件：

```yaml
processes:
  # 消息队列消费者
  - name: consumer
    command: ["/app/consumer"]
    health_check:
      type: tcp
      endpoint: localhost:9000
  
  # 数据处理器
  - name: processor
    command: ["/app/processor"]
    depends_on:
      - consumer
  
  # 结果上传器
  - name: uploader
    command: ["/app/uploader"]
    depends_on:
      - processor
```

### 4. 开发和测试环境

快速搭建本地开发环境：

```yaml
processes:
  # 数据库
  - name: postgres
    command: ["postgres", "-D", "/var/lib/postgresql/data"]
  
  # Redis 缓存
  - name: redis
    command: ["redis-server"]
  
  # 应用服务器
  - name: app
    command: ["npm", "run", "dev"]
    depends_on:
      - postgres
      - redis
```

## 🔧 故障排查

### 常见问题

#### 1. 进程无法启动

```bash
# 检查配置文件语法
guardian -config guardian.yaml --validate

# 查看详细日志
./guardian -config guardian.yaml -log-level debug

# 检查进程依赖
curl http://localhost:8080/api/processes | jq '.[] | {name, state, failure_count}'
```

#### 2. 健康检查失败

```bash
# 手动测试健康检查端点
curl -v http://localhost:80/healthz

# 查看健康检查指标
curl http://localhost:9090/metrics | grep guardian_health_check
```

#### 3. 僵尸进程问题

```bash
# 查看僵尸进程重启次数
curl http://localhost:8080/api/processes | jq '.[] | {name, zombie_restarts}'

# 检查系统僵尸进程
ps aux | awk '$8 ~ /Z/ {print $0}'
```

#### 4. 资源限制不生效

```bash
# 确认 Cgroup v2 挂载
mount | grep cgroup

# 检查 Cgroup 配置
cat /sys/fs/cgroup/cgroup.controllers

# 查看 Guardian 日志中的资源控制信息
grep "resource" /var/log/guardian/*.log
```

### 调试技巧

1. **启用调试模式**
   ```yaml
   global:
     debug: true
     log_level: debug
   ```

2. **查看实时日志**
   ```bash
   tail -f /var/log/guardian/*.log
   ```

3. **监控指标**
   ```bash
   watch -n 1 'curl -s http://localhost:9090/metrics | grep guardian_process_state'
   ```

4. **使用 API 调试**
   ```bash
   # 持续监控进程状态
   while true; do
     curl -s http://localhost:8080/api/processes | jq '.[] | {name, state, pid}'
     sleep 2
   done
   ```

## 🔄 进程状态机

```
┌─────────┐
│ stopped │◄────── 初始状态
└────┬────┘
     │ Start()
     ▼
┌──────────┐
│ starting │
└────┬─────┘
     │ Cmd.Start()
     ▼
┌─────────┐     健康检查连续失败      ┌───────┐
│ running │──────────────────────────►│ failed │
└────┬────┘                          └───┬───┘
     │                                    │
     │         健康检查恢复               │
     │◄──────────────────────────────────┘
     │
     │ Stop() / 进程退出
     ▼
┌──────────┐
│ stopping │
└──────────┘
```

### 状态转换说明

| 当前状态 | 触发事件 | 目标状态 | 说明 |
|----------|----------|----------|------|
| stopped | Start() | starting | 开始启动进程 |
| starting | Cmd.Start() 成功 | running | 进程启动成功 |
| starting | Cmd.Start() 失败 | failed | 启动失败 |
| running | 健康检查连续失败 | failed | 健康检查失败 |
| failed | 健康检查恢复 | running | 健康检查恢复 |
| running | Stop() | stopping | 开始停止进程 |
| running | 进程退出 | stopped | 进程正常退出 |
| failed | 重启策略触发 | starting | 尝试重启 |

## 📡 信号处理

| 信号 | 动作 |
|------|------|
| SIGINT | 优雅关闭（30 秒超时） |
| SIGTERM | 优雅关闭（30 秒超时） |
| SIGHUP | 重新加载配置（可选） |
| SIGUSR1 | 输出当前状态到日志 |

Guardian 接收信号后，并行优雅停止所有管理的进程，每个进程有可配置的超时时间。

### 优雅关闭流程

```
1. 接收 SIGTERM/SIGINT 信号
2. 停止接收新请求
3. 并行发送 SIGTERM 给所有子进程
4. 等待子进程退出（最多 30 秒）
5. 对未退出的进程发送 SIGKILL
6. 清理资源，退出
```

### 自定义超时时间

```yaml
global:
  shutdown_timeout: 60s  # 自定义关闭超时时间
```

## ❓ 为什么选择 Guardian？

### 横向对比

| 特性 | systemd | supervisord | Tini | **Guardian** |
|------|---------|-------------|------|--------------|
| 容器原生设计 | ❌ | ❌ | ✅ | **✅** |
| 单二进制文件 | ❌ | ❌ | ✅ (静态) | **✅** |
| 二进制大小 | ~50MB+ | ~5MB | ~300KB | **~15MB** |
| 零 Shell 依赖 | ❌ | ❌ | ✅ | **✅** |
| Prometheus 指标 | 需配置 | ❌ | ❌ | **✅ 内置** |
| 健康检查 | 有限 | 有限 | ❌ | **✅ 完整** |
| HTTP/TCP/Exec | 有限 | 有限 | ❌ | **✅ 全部支持** |
| 日志轮转 | 外部 | 外部 | ❌ | **✅ 内置** |
| Cgroup 支持 | ✅ | ❌ | ❌ | **✅** |
| 进程依赖管理 | ✅ | ✅ | ❌ | **✅** |
| 优雅关闭 | ✅ | ✅ | ❌ | **✅** |
| JSON 结构化日志 | 需配置 | ❌ | ❌ | **✅** |
| 僵尸进程检测 | ✅ | ❌ | ✅ | **✅ 自动处理** |
| RESTful API | ❌ | ❌ | ❌ | **✅** |
| 配置文件 | 复杂 | ini 格式 | 无 | **YAML** |

### 各工具适用场景

| 工具 | 最佳场景 | 局限 |
|------|----------|------|
| **systemd** | 传统 VM/物理机 | 体积大、复杂度高、不适合容器 |
| **supervisord** | 传统 Web 应用 | 功能有限、缺乏现代特性 |
| **Tini** | 纯进程收养 | 仅处理僵尸进程、无管理功能 |
| **Guardian** | **容器化微服务** | 新项目、生态仍在建设 |

### Tini 深度对比

Tini 是 Docker 官方推荐的轻量级 init 进程，主要功能是收养孤儿进程和处理僵尸进程。

| 功能 | Tini | **Guardian** |
|------|------|--------------|
| 进程收养 | ✅ | ✅ |
| 僵尸回收 | ✅ | ✅ |
| 进程管理 | ❌ | **✅** |
| 健康检查 | ❌ | **✅** |
| 自动重启 | ❌ | **✅** |
| Prometheus 指标 | ❌ | **✅** |
| 日志管理 | ❌ | **✅** |
| 资源限制 | ❌ | **✅** |
| 优雅关闭 | ❌ | **✅** |
| 进程依赖 | ❌ | **✅** |
| RESTful API | ❌ | **✅** |

**总结**：
- **Tini** = 极简 init，仅解决僵尸进程问题（~300KB）
- **Guardian** = 完整进程生命周期管理，包括监控、重启、日志、指标（~15MB）

如果只需要处理僵尸进程，用 **Tini**；如果需要完整的进程管理方案，用 **Guardian**。

## 进程管理 vs Tini

```
Tini: 仅处理进程收养
┌─────────────────────────────────────┐
│  PID 1 (Tini)                       │
│    └── 收养孤儿进程                  │
│    └── 回收僵尸进程                  │
│                                     │
│  注意：Tini 不启动/管理任何进程       │
└─────────────────────────────────────┘

Guardian: 完整生命周期管理
┌─────────────────────────────────────┐
│  PID 1 (Guardian)                   │
│    ├── 启动所有配置的进程            │
│    ├── 监控进程状态                  │
│    ├── 健康检查 + 自动重启            │
│    ├── 收集日志 + 轮转               │
│    ├── 导出 Prometheus 指标          │
│    ├── 应用 Cgroup 资源限制          │
│    ├── 处理进程依赖                  │
│    ├── 优雅关闭                      │
│    └── 回收僵尸进程（via init）       │
└─────────────────────────────────────┘
```

## 📚 开发指南

```bash
# 安装依赖
make install-deps

# 运行测试
make test

# 代码格式化
make fmt

# 运行 linter
make lint

# 编译二进制
make build

# 运行测试覆盖
make test
# 查看覆盖率报告
open coverage.html

# 清理构建产物
make clean
```

### 本地开发

```bash
# 1. 克隆仓库
git clone https://github.com/your-org/guardian.git
cd guardian

# 2. 安装 Go 依赖
go mod download

# 3. 运行 Guardian
go run cmd/guardian/main.go -config guardian.yaml

# 4. 运行测试
go test -v ./...

# 5. 构建并运行
make build
./guardian -config guardian.yaml
```

### 贡献代码

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 运行 `golangci-lint` 检查代码质量
- 为新功能添加测试用例

### 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v ./internal/process -run TestProcessManager

# 运行基准测试
go test -bench=. ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 贡献

我们欢迎各种形式的贡献！

### 如何贡献

1. 查看 [Issues](https://github.com/your-org/guardian/issues) 寻找可以帮忙的问题
2. Fork 仓库并创建分支
3. 实现功能或修复 bug
4. 添加测试确保质量
5. 提交 Pull Request

### 贡献者权益

- 在 README 中列出贡献者名单
- 参与项目决策讨论
- 获得社区认可

## 📄 License

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

感谢以下开源项目：

- [Prometheus](https://prometheus.io/) - 监控系统
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - YAML 解析
- [Docker](https://www.docker.com/) - 容器技术

## 📞 联系方式

- **Issues**: [GitHub Issues](https://github.com/your-org/guardian/issues)
- **Email**: your-email@example.com
- **Discussion**: [GitHub Discussions](https://github.com/your-org/guardian/discussions)

## 🔮 路线图

### v0.1.0 (当前版本)
- ✅ 基础进程管理
- ✅ 健康检查
- ✅ 日志管理
- ✅ Prometheus 指标
- ✅ RESTful API

### v0.2.0 (计划中)
- [ ] Web UI 管理界面
- [ ] 配置热重载
- [ ] 插件系统
- [ ] 分布式追踪支持

### v1.0.0 (未来)
- [ ] 生产环境稳定性验证
- [ ] 更多 Cgroup v2 功能
- [ ] 性能优化
- [ ] 完整文档

---

<div align="center">

**Made with ❤️ by the Guardian Team**

如果这个项目对你有帮助，请考虑给我们一个 ⭐️ Star！

[返回顶部](#guardian)

</div>

## 🏗️ 项目结构

```
guardian/
├── cmd/guardian/
│   └── main.go              # 程序入口
├── internal/
│   ├── config/              # YAML 配置加载
│   │   └── config.go        # 配置结构定义
│   ├── log/                 # 日志引擎（JSON + 轮转）
│   │   ├── engine.go        # 日志引擎核心
│   │   └── logger.go        # 日志记录器
│   ├── metrics/             # Prometheus 服务器
│   │   └── server.go        # Metrics HTTP 服务
│   ├── process/             # 进程生命周期管理
│   │   ├── manager.go       # 进程管理器
│   │   ├── manager_unix.go  # Unix 特定实现
│   │   ├── manager_windows.go # Windows 特定实现
│   │   ├── process.go       # 进程结构定义
│   │   ├── process_unix.go  # Unix 进程操作
│   │   └── process_windows.go # Windows 进程操作
│   ├── resource/            # Cgroup v2 资源控制
│   │   └── controller.go    # 资源控制器
│   ├── signal/              # 信号路由与优雅关闭
│   │   └── router.go        # 信号路由器
│   └── api/                 # RESTful API 服务
│       └── server.go        # API HTTP 服务
├── Dockerfile               # Docker 镜像构建
├── Makefile                 # 构建自动化
├── guardian.yaml            # 示例配置
├── go.mod                   # Go 模块定义
├── go.sum                   # 依赖校验
├── API.md                   # 详细 API 文档
├── test_api.sh              # API 测试脚本
└── test_zombie.sh           # 僵尸进程测试脚本
```

### 核心模块说明

| 模块 | 职责 | 关键文件 |
|------|------|----------|
| **cmd/guardian** | 程序入口，参数解析 | main.go |
| **config** | YAML 配置加载和验证 | config.go |
| **process** | 进程生命周期管理 | manager.go, process.go |
| **log** | 日志收集、轮转、JSON 格式化 | engine.go |
| **metrics** | Prometheus 指标收集和暴露 | server.go |
| **resource** | Cgroup v2 资源限制 | controller.go |
| **signal** | 信号处理和优雅关闭 | router.go |
| **api** | RESTful API 服务 | server.go |
