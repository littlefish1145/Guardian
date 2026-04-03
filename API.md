# Guardian API 文档

## API 端点

### 1. 获取所有进程信息

**端点:** `GET /api/processes`

**描述:** 获取所有进程的详细信息（格式化JSON输出）

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
  },
  {
    "name": "zombie-test",
    "state": "reclaimed",
    "pid": 16,
    "restarts": 3,
    "memory_mb": 7.85,
    "last_start": "2026-04-03T15:04:10Z",
    "failure_count": 5,
    "healthy_count": 0,
    "zombie_restarts": 3,
    "abandoned": true
  }
]
```

### 2. 获取单个进程信息

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

### 3. 健康检查

**端点:** `GET /api/health`

**描述:** API服务健康检查

**响应示例:**
```json
{
  "status": "ok",
  "time": "2026-04-03T15:04:05Z"
}
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

## 配置说明

### API 端口配置

在 `guardian.yaml` 中配置 API 端口：

```yaml
global:
  api_port: 8080  # API服务端口，默认8080
```

### 健康检查无上限

设置 `failure_threshold: -1` 表示健康检查失败无上限，不会触发重启：

```yaml
health_check:
  type: http
  endpoint: http://localhost:80/healthz
  interval: 10s
  timeout: 3s
  failure_threshold: -1  # -1 表示无上限
```

## 使用示例

### 使用 curl 查询所有进程

```bash
curl http://localhost:8080/api/processes
```

### 使用 curl 查询单个进程

```bash
curl http://localhost:8080/api/process/web-server
```

### 使用 curl 健康检查

```bash
curl http://localhost:8080/api/health
```

## 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 进程名称 |
| `state` | string | 进程状态 |
| `pid` | int | 进程ID |
| `restarts` | int | 重启次数 |
| `memory_mb` | float | 内存使用（MB） |
| `last_start` | string | 最后启动时间 |
| `failure_count` | int | 健康检查失败次数 |
| `healthy_count` | int | 健康检查成功次数 |
| `zombie_restarts` | int | 僵尸进程重启次数 |
| `abandoned` | bool | 是否已放弃重启 |
