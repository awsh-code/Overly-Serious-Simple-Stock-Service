# Monitoring Quick Reference

## Grafana Access
- **URL**: http://localhost:3000
- **Username**: admin  
- **Password**: 2hKJ4YDzGPNfIbDYJaf1mCeHvJvWL5d5MMy9TpF6%

## Prometheus Access
- **URL**: http://localhost:9090
- **Query Interface**: Available at root path

## Key Metrics to Monitor

### Application Metrics
```promql
# Request rate
rate(ping_service_requests_total[5m])

# Request duration (95th percentile)
histogram_quantile(0.95, rate(ping_service_request_duration_seconds_bucket[5m]))

# Error rate
rate(ping_service_errors_total[5m])

# Cache hit rate
rate(ping_service_cache_hits_total[5m]) / rate(ping_service_cache_requests_total[5m])
```

### Circuit Breaker Metrics
```promql
# Circuit breaker state (0=closed, 1=open, 2=half-open)
ping_service_circuit_breaker_state

# Circuit breaker failures
rate(ping_service_circuit_breaker_failures_total[5m])

# Circuit breaker successes  
rate(ping_service_circuit_breaker_successes_total[5m])
```

### Infrastructure Metrics
```promql
# CPU usage
rate(container_cpu_usage_seconds_total{pod=~"ping-service-.*"}[5m])

# Memory usage
container_memory_usage_bytes{pod=~"ping-service-.*"}

# Pod count
kube_deployment_status_replicas{deployment="ping-service"}
```

### HPA Metrics
```promql
# Current CPU utilization
kube_hpa_status_current_cpu_percentage{hpa="ping-service-hpa"}

# Current memory utilization  
kube_hpa_status_current_memory_percentage{hpa="ping-service-hpa"}

# Desired replicas
kube_hpa_status_desired_replicas{hpa="ping-service-hpa"}
```

## Alerting Rules (Example)
```yaml
groups:
- name: ping-service-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(ping_service_errors_total[5m]) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      
  - alert: CircuitBreakerOpen
    expr: ping_service_circuit_breaker_state > 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Circuit breaker is open"
      
  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(ping_service_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency detected"
```

## Dashboard URLs
- **Grafana**: http://localhost:3000/d/ping-service-overview
- **Prometheus**: http://localhost:9090/graph?g0.expr=ping_service_requests_total

## Troubleshooting Commands
```bash
# Check pod logs
kubectl logs -n dev -l app=ping-service --tail=50

# Check HPA status
kubectl describe hpa -n dev ping-service-hpa

# Check metrics endpoint
curl -s http://ping-service.46.225.33.158.nip.io/metrics | grep ping_service

# Port forward for local access
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
kubectl port-forward -n monitoring service/prometheus-grafana 3000:80
```