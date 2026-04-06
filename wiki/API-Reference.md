# API 文档

Guardian 提供 RESTful API 用于查询进程状态和管理。

## 基础信息

- **API 地址**: `http://localhost:8080`
- **数据格式**: JSON
- **认证**: 暂无（计划中）

## API 端点

### 1. 获取所有进程信息

**端点**: `GET /api/processes`

**描述**: 获取所有进程的详细信息（格式化 JSON 输出）

**响应示例**:
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

### 2. 获取单个进程信息

**端点**: `GET /api/process/{name}`

**描述**: 根据进程名获取单个进程的详细信息

**参数**:
- `name` - 进程名称

**响应示例**:
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

### 3. 健康检查

**端点**: `GET /api/health`

**描述**: API 服务健康检查

**响应示例**:
```json
{
  "status": "ok",
  "time": "2026-04-03T15:04:05Z"
}
```

## 使用示例

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

## 进程状态说明

| 状态 | 说明 |
|------|------|
| `starting` | 进程启动中 |
| `running` | 进程运行中 |
| `stopping` | 进程停止中 |
| `stopped` | 进程已停止 |
| `failed` | 进程失败 |
| `reclaimed` | 进程已回收（僵尸进程专用） |

## 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 进程名称 |
| `state` | string | 进程状态（见上表） |
| `pid` | int | 进程 ID |
| `restarts` | int | 重启次数 |
| `memory_mb` | float | 内存使用（MB） |
| `last_start` | string | 最后启动时间（ISO 8601） |
| `failure_count` | int | 健康检查失败次数 |
| `healthy_count` | int | 健康检查成功次数 |
| `zombie_restarts` | int | 僵尸进程重启次数 |
| `abandoned` | bool | 是否被放弃（超过最大重启次数） |

## 错误响应

### 404 Not Found

当查询的进程不存在时：

```json
{
  "error": "Process 'web-server' not found"
}
```

### 500 Internal Server Error

服务器内部错误：

```json
{
  "error": "Internal server error"
}
```

## Prometheus 指标

Guardian 在 `http://localhost:9090/metrics` 暴露 Prometheus 指标。

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

### 其他端点

- `/ready` - 就绪探针
- `/live` - 存活探针
- `/api/health` - API 健康检查

## 相关文档

- [[配置指南 | Configuration]]
- [[快速开始 | Getting-Started]]
