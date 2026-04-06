# Kubernetes 部署

Guardian 可作为 init 容器或 sidecar 运行在 Kubernetes 环境中。

> **注意**: 在 K8s 环境中，我们建议使用原生的 init 容器或 sidecar 模式。当然，您仍可以使用 Guardian 来管理多个应用进程。

## 作为 Init 容器

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

## 作为 Sidecar 容器

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

## Deployment 示例

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

## 配置 ConfigMap

```bash
# 创建 ConfigMap
kubectl create configmap guardian-config --from-file=guardian.yaml

# 或使用 YAML
apiVersion: v1
kind: ConfigMap
metadata:
  name: guardian-config
data:
  guardian.yaml: |
    global:
      log_level: info
      metrics_port: 9090
      api_port: 8080
    processes:
      - name: app
        command: ["/app/server"]
        health_check:
          type: http
          endpoint: http://localhost:8080/health
```

## ServiceMonitor（Prometheus Operator）

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

## 权限要求

Guardian 需要 `SYS_ADMIN` 权限来管理 Cgroup：

```yaml
securityContext:
  capabilities:
    add: ["SYS_ADMIN"]
```

> **安全提示**: 在生产环境中，请谨慎授予权限，考虑使用 PodSecurityPolicy 或 Pod Security Standards 限制权限范围。

## 资源配额

建议为 Guardian 配置资源请求和限制：

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## 健康检查

配置 K8s 探针监控 Guardian 状态：

```yaml
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
```

## 监控和日志

### 查看 Pod 状态

```bash
kubectl get pods -l app=guardian
kubectl describe pod <pod-name>
```

### 查看日志

```bash
kubectl logs <pod-name> -c guardian
kubectl logs -f <pod-name> -c guardian
```

### 访问 Prometheus 指标

```bash
kubectl port-forward <pod-name> 9090:9090
# 然后访问 http://localhost:9090/metrics
```

## 相关文档

- [[Docker 使用 | Docker-Usage]]
- [[配置指南 | Configuration]]
