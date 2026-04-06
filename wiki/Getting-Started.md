# 快速开始

本指南将帮助你快速上手使用 Guardian。

## 环境要求

- **Go** 1.21+
- **Docker** 20.10+（可选）
- **Make**（可选，方便构建）
- **Linux** 环境（推荐，完整支持 Cgroup v2）

## 编译

### 本地编译

```bash
# 克隆仓库
git clone https://github.com/your-org/guardian.git
cd guardian

# 安装依赖
make install-deps

# 本地编译
make build

# 交叉编译（其他平台）
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o guardian
```

### Docker 编译

```bash
make docker-build
```

## 运行

### 本地运行

```bash
# 使用默认配置运行
./guardian -config guardian.yaml

# 指定日志级别
./guardian -config guardian.yaml -log-level debug
```

### Docker 运行

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

### Kubernetes 部署

```bash
# 创建 ConfigMap
kubectl create configmap guardian-config --from-file=guardian.yaml

# 部署 Pod
kubectl apply -f k8s/deployment.yaml
```

## 快速配置示例

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

## 验证运行

### 检查进程状态

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

### 查看 Prometheus 指标

访问 `http://localhost:9090/metrics` 查看监控指标。

## 下一步

- 查看 [[配置指南 | Configuration]] 了解详细配置选项
- 查看 [[API 文档 | API-Reference]] 了解 API 接口
- 查看 [[Docker 使用 | Docker-Usage]] 了解容器化部署
