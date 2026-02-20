# Deployment Guide

This guide covers deploying the Overly Serious Simple Stock Service to Kubernetes using both raw manifests and Helm charts.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment Options](#deployment-options)
- [Kubernetes Manifests](#kubernetes-manifests)
- [Helm Chart](#helm-chart)
- [Monitoring Setup](#monitoring-setup)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured to access your cluster
- For Helm deployment: Helm 3.2.0+
- For monitoring: Prometheus Operator installed
- Container registry access for pulling images

## Deployment Options

### Option 1: Kubernetes Manifests (Development/Simple)

Best for development environments or when you want full control over every resource.

```bash
# Apply all manifests
kubectl apply -f k8s/

# Or apply individually
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secret.yaml      # Set your API key first
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/ingress.yaml     # Optional
kubectl apply -f k8s/hpa.yaml          # Optional
kubectl apply -f k8s/servicemonitor.yaml  # Optional
```

### Option 2: Helm Chart (Production/Recommended)

Best for production environments with comprehensive configuration options.

#### Basic Installation

```bash
# Quick start with minimal configuration
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --create-namespace \
  --set config.stockAPI.apiKey=YOUR_API_KEY
```

#### Production Installation

```bash
# Production-ready installation with all features
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --create-namespace \
  --set config.stockAPI.apiKey=YOUR_API_KEY \
  --set replicaCount=3 \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=3 \
  --set autoscaling.maxReplicas=10 \
  --set monitoring.enabled=true \
  --set monitoring.serviceMonitor.enabled=true \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=stock-service.yourdomain.com \
  --set config.stockAPI.circuitBreaker.enabled=true \
  --set config.stockAPI.cache.enabled=true \
  --set resources.limits.memory=512Mi \
  --set resources.limits.cpu=500m \
  --set resources.requests.memory=256Mi \
  --set resources.requests.cpu=250m
```

#### Custom Values File

Create a `values-production.yaml` file:

```yaml
# Production values
replicaCount: 3

config:
  stockAPI:
    apiKey: YOUR_API_KEY  # Replace with your API key
    circuitBreaker:
      enabled: true
      failureThreshold: 5
      timeout: 30s
    cache:
      enabled: true
      ttl: 300s

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: stock-service.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: stock-service-tls
      hosts:
        - stock-service.yourdomain.com

resources:
  limits:
    memory: 512Mi
    cpu: 500m
  requests:
    memory: 256Mi
    cpu: 250m

security:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

Deploy with custom values:

```bash
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --create-namespace \
  -f values-production.yaml
```

## Kubernetes Manifests

The raw Kubernetes manifests provide a straightforward deployment option:

### Namespace
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: stock-service
  labels:
    name: stock-service
```

### Secret (API Key)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: stock-service-secret
  namespace: stock-service
type: Opaque
data:
  stock-api-key: YOUR_BASE64_ENCODED_API_KEY
```

### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stock-service
  namespace: stock-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: stock-service
  template:
    metadata:
      labels:
        app: stock-service
    spec:
      containers:
      - name: stock-service
        image: ghcr.io/your-org/stock-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: STOCK_API_KEY
          valueFrom:
            secretKeyRef:
              name: stock-service-secret
              key: stock-api-key
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
          requests:
            memory: "256Mi"
            cpu: "250m"
```

### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: stock-service
  namespace: stock-service
spec:
  selector:
    app: stock-service
  ports:
  - name: http
    port: 80
    targetPort: 8080
  type: ClusterIP
```

## Helm Chart

The Helm chart provides comprehensive configuration options:

### Chart Structure
```
charts/stock-service/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default values
├── templates/              # Kubernetes templates
│   ├── _helpers.tpl        # Template helpers
│   ├── deployment.yaml     # Main application
│   ├── service.yaml        # Service definition
│   ├── ingress.yaml        # Ingress configuration
│   ├── hpa.yaml           # Horizontal Pod Autoscaler
│   ├── servicemonitor.yaml # Prometheus monitoring
│   ├── secret.yaml         # API key secret
│   └── serviceaccount.yaml # Service account
└── README.md              # Chart documentation
```

### Key Features

- **Autoscaling**: CPU and memory-based horizontal pod autoscaling
- **Monitoring**: Prometheus ServiceMonitor for metrics collection
- **Security**: Non-root containers, read-only filesystem, resource limits
- **Configuration**: Comprehensive configuration via values.yaml
- **Ingress**: NGINX ingress with TLS support
- **Circuit Breaker**: Built-in circuit breaker for external API calls
- **Caching**: In-memory caching with configurable TTL

## Monitoring Setup

### Prometheus Integration

1. **Install Prometheus Operator** (if not already installed):
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

2. **Enable ServiceMonitor**:
```bash
helm upgrade stock-service ./charts/stock-service \
  --set monitoring.serviceMonitor.enabled=true
```

### Grafana Dashboards

Import the provided dashboards:

1. **Golden Signals Dashboard** (`monitoring/dashboards/golden-signals.json`)
2. **Stock Service Dashboard** (`monitoring/dashboards/stock-service.json`)

### Key Metrics to Monitor

- **Request Rate**: `rate(http_requests_total[5m])`
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **Latency**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Circuit Breaker State**: `stock_service_circuit_breaker_state`
- **Cache Hit Rate**: `rate(stock_service_cache_hits_total[5m]) / (rate(stock_service_cache_hits_total[5m]) + rate(stock_service_cache_misses_total[5m]))`

## Security Considerations

### API Key Management

1. **Never commit API keys to Git**
2. **Use Kubernetes secrets or external secret management**
3. **Rotate API keys regularly**
4. **Use least-privilege API keys**

### Network Security

1. **Use NetworkPolicies to restrict traffic**
2. **Enable TLS for all external communications**
3. **Use ingress controllers with security features**
4. **Implement rate limiting**

### Container Security

1. **Use non-root containers**
2. **Read-only root filesystem**
3. **Drop unnecessary capabilities**
4. **Regular security scanning**
5. **Use distroless or minimal base images**

## Troubleshooting

### Common Issues

#### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n stock-service

# Check pod logs
kubectl logs -n stock-service deployment/stock-service

# Describe pod for events
kubectl describe pod -n stock-service -l app.kubernetes.io/name=stock-service
```

#### Service Not Accessible

```bash
# Check service status
kubectl get svc -n stock-service

# Port forward to test locally
kubectl port-forward -n stock-service svc/stock-service 8080:8080

# Test health endpoint
curl http://localhost:8080/health
```

#### High Memory Usage

```bash
# Check resource usage
kubectl top pods -n stock-service

# Check memory metrics in Grafana
# Look for: container_memory_usage_bytes

# Adjust resource limits
helm upgrade stock-service ./charts/stock-service \
  --set resources.limits.memory=1Gi
```

#### Circuit Breaker Triggered

```bash
# Check circuit breaker metrics
curl http://localhost:8080/metrics | grep circuit_breaker

# Check external API connectivity
kubectl exec -n stock-service deployment/stock-service -- \
  curl -I https://api.polygon.io/v1/marketstatus/now
```

### Debug Commands

```bash
# Get all resources in namespace
kubectl get all -n stock-service

# Check events
kubectl get events -n stock-service --sort-by='.lastTimestamp'

# Shell into pod for debugging
kubectl exec -n stock-service deployment/stock-service -it -- /bin/sh

# Check service endpoints
kubectl get endpoints -n stock-service

# View ingress status
kubectl get ingress -n stock-service
```

### Performance Tuning

1. **Adjust cache TTL** based on your data freshness requirements
2. **Tune circuit breaker thresholds** based on external API reliability
3. **Set appropriate resource requests/limits** based on actual usage
4. **Configure autoscaling parameters** based on traffic patterns
5. **Optimize database queries** if using persistent storage

## Upgrade Strategies

### Rolling Updates (Default)

```bash
# Update image version
helm upgrade stock-service ./charts/stock-service \
  --set image.tag=new-version
```

### Blue-Green Deployment

```bash
# Deploy to new namespace (green)
helm install stock-service-green ./charts/stock-service \
  --namespace stock-service-green \
  --set config.stockAPI.apiKey=YOUR_API_KEY

# Switch traffic (update ingress or service selector)
# Then delete old deployment
helm uninstall stock-service --namespace stock-service
```

### Canary Deployment

Use Flagger or similar tools for automated canary deployments based on metrics.

## Cleanup

### Remove Helm Release

```bash
helm uninstall stock-service --namespace stock-service
kubectl delete namespace stock-service
```

### Remove Raw Manifests

```bash
kubectl delete -f k8s/
```

This completes the comprehensive deployment guide covering both Kubernetes manifests and Helm chart deployment options with extensive monitoring, security, and troubleshooting information.