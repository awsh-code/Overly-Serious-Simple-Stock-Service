# Stock Service Helm Chart

This Helm chart deploys the Overly Serious Simple Stock Service to Kubernetes clusters.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Prometheus Operator (optional, for ServiceMonitor)

## Installation

### Quick Start

```bash
# Add your stock API key
helm install stock-service ./charts/stock-service \
  --set config.stockAPI.apiKey=YOUR_API_KEY
```

### Production Installation

```bash
# Create namespace
kubectl create namespace stock-service

# Install with production values
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --set config.stockAPI.apiKey=YOUR_API_KEY \
  --set replicaCount=3 \
  --set autoscaling.enabled=true \
  --set monitoring.enabled=true \
  --set monitoring.serviceMonitor.enabled=true \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=stock-service.yourdomain.com
```

## Configuration

### Required Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.stockAPI.apiKey` | API key for stock data service | `""` |

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `autoscaling.enabled` | Enable horizontal pod autoscaling | `false` |
| `autoscaling.minReplicas` | Minimum number of replicas | `1` |
| `autoscaling.maxReplicas` | Maximum number of replicas | `10` |
| `monitoring.enabled` | Enable Prometheus monitoring | `true` |
| `monitoring.serviceMonitor.enabled` | Enable ServiceMonitor | `false` |
| `ingress.enabled` | Enable ingress | `false` |
| `config.stockAPI.circuitBreaker.enabled` | Enable circuit breaker | `true` |
| `config.stockAPI.cache.enabled` | Enable caching | `true` |

### Security Configuration

The chart includes several security best practices:

- Non-root container execution
- Read-only root filesystem
- Dropped capabilities
- Resource limits and requests
- Network policies (optional)

## Monitoring

### Prometheus Metrics

The service exposes metrics at `/metrics` endpoint. Key metrics include:

- `ping_service_stock_api_duration_seconds`: Stock API request duration
- `ping_service_cache_hits_total`: Total cache hits
- `ping_service_cache_misses_total`: Total cache misses
- `ping_service_circuit_breaker_state`: Current circuit breaker state
- Standard HTTP metrics (request count, duration, etc.)

### Grafana Dashboards

Two pre-configured dashboards are available:

1. **Golden Signals Dashboard**: Focuses on the four golden signals (latency, traffic, errors, saturation)
2. **Ping Service Dashboard**: Detailed service-specific metrics including cache performance and circuit breaker state

## Upgrading

```bash
# Upgrade to a new version
helm upgrade stock-service ./charts/stock-service \
  --namespace stock-service \
  --set config.stockAPI.apiKey=YOUR_API_KEY
```

## Uninstallation

```bash
helm uninstall stock-service --namespace stock-service
```

## Troubleshooting

### Check Pod Status
```bash
kubectl get pods -n stock-service -l app.kubernetes.io/name=stock-service
```

### View Logs
```bash
kubectl logs -n stock-service -l app.kubernetes.io/name=stock-service -f
```

### Check Service Health
```bash
kubectl port-forward -n stock-service svc/stock-service 8080:8080
curl http://localhost:8080/health
```

### Verify Metrics
```bash
curl http://localhost:8080/metrics | grep ping_service
```

## Development

### Linting the Chart
```bash
helm lint ./charts/stock-service
```

### Template Rendering
```bash
helm template stock-service ./charts/stock-service \
  --set config.stockAPI.apiKey=test-key
```

### Dry Run Installation
```bash
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --set config.stockAPI.apiKey=test-key \
  --dry-run --debug
```