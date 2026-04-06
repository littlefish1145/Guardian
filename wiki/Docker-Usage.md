# Docker 使用

Guardian 专为容器 PID 1 场景设计，完美适配 Docker 环境。

## Dockerfile 示例

### 基础示例

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

## Docker Compose 示例

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

## 运行容器

### 构建镜像

```bash
make docker-build
```

或使用 docker 命令：

```bash
docker build -t guardian:latest .
```

### 启动容器

```bash
docker run --rm -it \
  -p 9090:9090 \
  -p 8080:8080 \
  -v $(pwd)/guardian.yaml:/etc/guardian/guardian.yaml \
  guardian:latest
```

### 后台运行

```bash
docker run -d \
  --name guardian \
  -p 9090:9090 \
  -p 8080:8080 \
  -v $(pwd)/guardian.yaml:/etc/guardian/guardian.yaml \
  -v $(pwd)/logs:/var/log/guardian \
  guardian:latest
```

### 查看日志

```bash
docker logs guardian
docker logs -f guardian  # 跟随日志输出
```

### 进入容器

```bash
docker exec -it guardian /bin/sh
```

## 容器配置说明

### 端口映射

- **9090**: Prometheus 指标端口
- **8080**: RESTful API 端口

### 卷挂载

- **配置文件**: 挂载 `guardian.yaml` 到 `/etc/guardian/guardian.yaml`
- **日志目录**: 挂载宿主机目录到 `/var/log/guardian`

### 环境变量

可通过环境变量覆盖配置文件中的设置：

```bash
docker run -e LOG_LEVEL=debug guardian:latest
```

## 健康检查

Docker Compose 中配置了健康检查：

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 最佳实践

1. **使用只读挂载**: 配置文件使用 `:ro` 只读挂载
2. **日志持久化**: 挂载日志目录到宿主机
3. **资源限制**: 使用 Docker 资源限制防止资源耗尽
4. **健康检查**: 配置 Docker 健康检查监控容器状态

## 相关文档

- [[Kubernetes 部署 | Kubernetes-Deployment]]
- [[配置指南 | Configuration]]
