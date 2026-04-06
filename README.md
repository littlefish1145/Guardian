# Guardian

<div align="center">

![Guardian Logo](./logo.png)

**云原生轻量级进程守护程序，专为容器环境设计**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-EA8139?style=flat&logo=prometheus)](https://prometheus.io/)

[快速开始](./wiki/Getting-Started) • [功能特性](#功能特性) • [配置指南](./wiki/Configuration) • [API 文档](./wiki/API-Reference) • [使用场景](#使用场景)

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

详细的使用指南请查看 [Wiki 文档](./wiki/Getting-Started)。

### 快速示例

```bash
# 编译
make build

# 运行
./guardian -config guardian.yaml
```

更多运行方式（Docker、Kubernetes）请查看 [快速开始指南](./wiki/Getting-Started)。

## 📝 配置指南

详细的配置说明请查看 [配置指南](./wiki/Configuration)。

### 配置概览

Guardian 使用 YAML 配置文件，支持以下主要配置项：

- **全局配置**: 日志级别、指标端口、API 端口等
- **进程配置**: 命令、依赖、环境变量等
- **健康检查**: HTTP/TCP/Exec 检查
- **重启策略**: always/on-failure/never，支持指数退避
- **资源限制**: CPU/内存/文件描述符限制
- **日志配置**: JSON 格式化、自动轮转

查看 [完整配置文档](./wiki/Configuration) 了解所有配置选项。

## 📊 API 文档

详细的 API 文档请查看 [API Reference](./wiki/API-Reference)。

### 快速示例

```bash
# 查询所有进程
curl http://localhost:8080/api/processes

# 查询单个进程
curl http://localhost:8080/api/process/web-server

# 健康检查
curl http://localhost:8080/api/health
```

### 主要端点

- `GET /api/processes` - 获取所有进程信息
- `GET /api/process/{name}` - 获取单个进程信息
- `GET /api/health` - API 健康检查

完整 API 文档请查看 [API Reference](./wiki/API-Reference)。

## 📊 Prometheus 指标

Guardian 在 `http://localhost:9090/metrics` 暴露 Prometheus 指标。

主要指标包括：

- `guardian_process_state` - 进程状态
- `guardian_process_restarts_total` - 重启次数
- `guardian_process_healthy` - 健康检查状态
- `guardian_process_uptime_seconds` - 运行时间
- `guardian_process_memory_bytes` - 内存使用

完整指标列表请查看 [API Reference](./wiki/API-Reference)。

## 🐳 Docker 使用

详细的 Docker 使用指南请查看 [Docker Usage](./wiki/Docker-Usage)。

### 快速示例

```dockerfile
FROM nginx:alpine

# 复制 Guardian
COPY --from=builder /build/guardian /usr/local/bin/
COPY guardian.yaml /etc/guardian/guardian.yaml

EXPOSE 9090 8080

ENTRYPOINT ["/usr/local/bin/guardian", "-config", "/etc/guardian/guardian.yaml"]
```

更多 Docker 和 Docker Compose 配置请查看 [Docker Usage](./wiki/Docker-Usage)。

## ☸️ Kubernetes 使用

详细的 K8s 部署指南请查看 [Kubernetes Deployment](./wiki/Kubernetes-Deployment)。

> **建议**: 在 K8s 环境中，我们建议使用原生的 init 容器或 sidecar 模式。当然，您仍可以使用 Guardian 来管理多个应用进程。

### 快速示例

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-with-guardian
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
    - name: app
      image: myapp:latest
```

更多 Deployment、ServiceMonitor 配置请查看 [Kubernetes Deployment](./wiki/Kubernetes-Deployment)。

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

遇到问题？请查看 [故障排查指南](./wiki/Troubleshooting)。

常见问题：

- 进程无法启动
- 健康检查失败
- 僵尸进程问题
- 资源限制不生效

详细解决方案请查看 [故障排查指南](./wiki/Troubleshooting)。

## 🔄 进程状态机

Guardian 进程状态转换流程：

```
stopped → starting → running → stopping → stopped
                    ↓
                  failed
```

状态说明：

- `starting` - 进程启动中
- `running` - 进程运行中
- `stopping` - 进程停止中
- `stopped` - 进程已停止
- `failed` - 进程失败
- `reclaimed` - 进程已回收（僵尸进程专用）

详细状态转换说明请查看 [API Reference](./wiki/API-Reference)。

## 📡 信号处理

| 信号 | 动作 |
|------|------|
| SIGINT | 优雅关闭（30 秒超时） |
| SIGTERM | 优雅关闭（30 秒超时） |
| SIGHUP | 重新加载配置（可选） |
| SIGUSR1 | 输出当前状态到日志 |

Guardian 接收信号后，并行优雅停止所有管理的进程，每个进程有可配置的超时时间。

### 优雅关闭流程

1. 接收 SIGTERM/SIGINT 信号
2. 停止接收新请求
3. 并行发送 SIGTERM 给所有子进程
4. 等待子进程退出（最多 30 秒）
5. 对未退出的进程发送 SIGKILL
6. 清理资源，退出

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

# 清理构建产物
make clean
```

### 本地开发

```bash
# 1. 克隆仓库
git clone https://github.com/littlefish1145/guardian.git
cd guardian

# 2. 运行 Guardian
go run cmd/guardian/main.go -config guardian.yaml

# 3. 运行测试
go test -v ./...
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

## 🤝 贡献

我们欢迎各种形式的贡献！

### 如何贡献

1. 查看 [Issues](https://github.com/littlefish/guardian/issues) 寻找可以帮忙的问题
2. Fork 仓库并创建分支
3. 实现功能或修复 bug
4. 添加测试确保质量
5. 提交 Pull Request

### 贡献者权益

还没有喵...只有我一个开发喵...

## 📄 License

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

感谢以下开源项目：

- [Prometheus](https://prometheus.io/) - 监控系统
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - YAML 解析
- [Docker](https://www.docker.com/) - 容器技术

## 📞 联系方式

- **Issues**: [GitHub Issues](https://github.com/littlefish1145/guardian/issues)
- **Email**: 3951168381@qq.com
- **Discussion**: [GitHub Discussions](https://github.com/littlefish1145/guardian/discussions)

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

## 📚 项目结构

```
guardian/
├── cmd/guardian/
│   └── main.go              # 程序入口
├── internal/
│   ├── config/              # YAML 配置加载
│   ├── log/                 # 日志引擎（JSON + 轮转）
│   ├── metrics/             # Prometheus 服务器
│   ├── process/             # 进程生命周期管理
│   ├── resource/            # Cgroup v2 资源控制
│   ├── signal/              # 信号路由与优雅关闭
│   └── api/                 # RESTful API 服务
├── Dockerfile               # Docker 镜像构建
├── Makefile                 # 构建自动化
├── guardian.yaml            # 示例配置
└── go.mod                   # Go 模块定义
```

### 核心模块说明

| 模块 | 职责 |
|------|------|
| **cmd/guardian** | 程序入口，参数解析 |
| **config** | YAML 配置加载和验证 |
| **process** | 进程生命周期管理 |
| **log** | 日志收集、轮转、JSON 格式化 |
| **metrics** | Prometheus 指标收集和暴露 |
| **resource** | Cgroup v2 资源限制 |
| **signal** | 信号处理和优雅关闭 |
| **api** | RESTful API 服务 |

---

<div align="center">

**Made with ❤️ by the Kimu**

如果这个项目对你有帮助，请考虑给我一个 ⭐️ Star！球球了喵~

[返回顶部](#guardian)

</div>
