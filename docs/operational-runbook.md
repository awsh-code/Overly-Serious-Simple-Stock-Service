# Operational Runbook - Stock Service

This runbook provides operational procedures for maintaining and troubleshooting the Overly Serious Simple Stock Service in production.

## Table of Contents

- [Service Overview](#service-overview)
- [Key Metrics and Alerts](#key-metrics-and-alerts)
- [Incident Response Procedures](#incident-response-procedures)
- [Scaling Procedures](#scaling-procedures)
- [Maintenance Procedures](#maintenance-procedures)
- [Disaster Recovery](#disaster-recovery)
- [Contact Information](#contact-information)

## Service Overview

### Service Details
- **Name**: Stock Service
- **Purpose**: Provides stock market data with caching and circuit breaker protection
- **Technology**: Go with Gin framework
- **Port**: 8080
- **Health Endpoint**: `/health`
- **Metrics Endpoint**: `/metrics`
- **External Dependency**: Polygon.io API

### Architecture
- **Deployment**: Kubernetes Deployment with HPA
- **Service**: ClusterIP service
- **Ingress**: NGINX Ingress Controller
- **Monitoring**: Prometheus + Grafana
- **Caching**: In-memory cache with TTL
- **Circuit Breaker**: Hystrix-style circuit breaker

## Key Metrics and Alerts

### Critical Alerts (P1 - Page Immediately)

#### Service Down
- **Alert**: `ServiceDown`
- **Condition**: `up{job="stock-service"} == 0`
- **Threshold**: 0
- **Duration**: 1 minute
- **Action**: Check pod status, logs, and external dependencies

#### High Error Rate
- **Alert**: `HighErrorRate`
- **Condition**: `rate(http_requests_total{status=~"5.."}[5m]) > 0.1`
- **Threshold**: 10% error rate
- **Duration**: 5 minutes
- **Action**: Check application logs and external API status

#### Circuit Breaker Open
- **Alert**: `CircuitBreakerOpen`
- **Condition**: `stock_service_circuit_breaker_state == 2`
- **Threshold**: 2 (Open state)
- **Duration**: Immediate
- **Action**: Check external API availability and response times

### Warning Alerts (P2 - Business Hours)

#### High Latency
- **Alert**: `HighLatency`
- **Condition**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2`
- **Threshold**: 2 seconds
- **Duration**: 10 minutes
- **Action**: Check cache hit rate and external API performance

#### High Memory Usage
- **Alert**: `HighMemoryUsage`
- **Condition**: `container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9`
- **Threshold**: 90% memory usage
- **Duration**: 15 minutes
- **Action**: Check for memory leaks or insufficient resources

#### Low Cache Hit Rate
- **Alert**: `LowCacheHitRate`
- **Condition**: `rate(stock_service_cache_hits_total[5m]) / (rate(stock_service_cache_hits_total[5m]) + rate(stock_service_cache_misses_total[5m])) < 0.5`
- **Threshold**: 50% hit rate
- **Duration**: 30 minutes
- **Action**: Review cache configuration and TTL settings

### Info Alerts (P3 - Next Business Day)

#### Deployment Events
- **Alert**: `DeploymentCompleted`
- **Condition**: `kube_deployment_status_replicas_updated == kube_deployment_spec_replicas`
- **Action**: Verify deployment success and run smoke tests

#### Certificate Expiry
- **Alert**: `CertificateExpiry`
- **Condition**: `certmanager_certificate_expiration_timestamp_seconds - time() < 86400 * 30`
- **Threshold**: 30 days
- **Action**: Schedule certificate renewal

## Incident Response Procedures

### P1 Incident Response (Service Down)

#### Immediate Actions (0-5 minutes)
1. **Acknowledge Alert**: Respond to page within 5 minutes
2. **Check Service Status**:
   ```bash
   kubectl get pods -n stock-service
   kubectl describe pod -n stock-service -l app.kubernetes.io/name=stock-service
   ```
3. **Check Logs**:
   ```bash
   kubectl logs -n stock-service -l app.kubernetes.io/name=stock-service --tail=100
   ```
4. **Check External Dependencies**:
   ```bash
   curl -I https://api.polygon.io/v1/marketstatus/now
   ```

#### Investigation (5-15 minutes)
1. **Check Circuit Breaker State**:
   ```bash
   kubectl port-forward -n stock-service svc/stock-service 8080:8080
   curl http://localhost:8080/metrics | grep circuit_breaker
   ```
2. **Check Resource Usage**:
   ```bash
   kubectl top pods -n stock-service
   ```
3. **Check Recent Deployments**:
   ```bash
   kubectl rollout history deployment/stock-service -n stock-service
   ```

#### Resolution Actions (15-30 minutes)
1. **If Pod Issues**: Restart deployment
   ```bash
   kubectl rollout restart deployment/stock-service -n stock-service
   ```
2. **If Resource Issues**: Scale up resources
   ```bash
   kubectl scale deployment/stock-service -n stock-service --replicas=5
   ```
3. **If External API Issues**: Verify circuit breaker behavior
4. **If Configuration Issues**: Rollback to previous version
   ```bash
   kubectl rollout undo deployment/stock-service -n stock-service
   ```

#### Communication
- **Internal**: Notify team via Slack #alerts channel
- **External**: Update status page if customer-facing
- **Stakeholders**: Send incident report within 1 hour

### P2 Incident Response (Performance Issues)

#### Investigation
1. **Check Metrics Dashboard**: Review Grafana dashboards
2. **Check Cache Performance**: 
   ```bash
   curl http://localhost:8080/metrics | grep cache
   ```
3. **Check External API Latency**:
   ```bash
   curl http://localhost:8080/metrics | grep stock_api_duration
   ```

#### Resolution
1. **High Latency**: Check cache hit rate and adjust TTL
2. **High Memory**: Review memory usage patterns and adjust limits
3. **Low Cache Hit Rate**: Analyze request patterns and cache configuration

## Scaling Procedures

### Manual Scaling

#### Scale Up
```bash
# Using kubectl
kubectl scale deployment/stock-service -n stock-service --replicas=10

# Using Helm
helm upgrade stock-service ./charts/stock-service \
  --namespace stock-service \
  --set replicaCount=10
```

#### Scale Down
```bash
# Using kubectl
kubectl scale deployment/stock-service -n stock-service --replicas=3

# Using Helm
helm upgrade stock-service ./charts/stock-service \
  --namespace stock-service \
  --set replicaCount=3
```

### Autoscaling Configuration

#### Enable HPA
```bash
helm upgrade stock-service ./charts/stock-service \
  --namespace stock-service \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=3 \
  --set autoscaling.maxReplicas=20 \
  --set autoscaling.targetCPUUtilizationPercentage=70
```

#### Monitor Autoscaling
```bash
kubectl get hpa -n stock-service -w
```

### Capacity Planning

#### Baseline Metrics
- **CPU**: 250m per pod (baseline)
- **Memory**: 256Mi per pod (baseline)
- **Traffic**: 1000 RPS per pod (baseline)

#### Scaling Triggers
- **CPU > 70%**: Scale up
- **Memory > 80%**: Scale up
- **P99 Latency > 2s**: Scale up
- **Error Rate > 5%**: Scale up

## Maintenance Procedures

### Regular Maintenance (Weekly)

#### Health Check
1. **Verify All Pods Running**:
   ```bash
   kubectl get pods -n stock-service
   ```
2. **Check Resource Usage**:
   ```bash
   kubectl top pods -n stock-service
   ```
3. **Review Logs for Errors**:
   ```bash
   kubectl logs -n stock-service -l app.kubernetes.io/name=stock-service --tail=1000 | grep ERROR
   ```

#### Performance Review
1. **Check Cache Performance**:
   - Review cache hit rate trends
   - Adjust TTL if necessary
2. **Review Circuit Breaker Activity**:
   - Check for frequent openings
   - Verify timeout settings
3. **Analyze Traffic Patterns**:
   - Review peak usage times
   - Plan capacity accordingly

### Monthly Maintenance

#### Security Updates
1. **Update Base Images**:
   ```bash
   # Update Dockerfile base image
   # Rebuild and deploy
   ```
2. **Review Dependencies**:
   - Check for security vulnerabilities
   - Update Go dependencies
3. **Certificate Renewal**:
   - Check TLS certificate expiry
   - Renew if necessary

#### Performance Optimization
1. **Review Resource Limits**:
   - Analyze actual usage vs. limits
   - Optimize for efficiency
2. **Tune Cache Settings**:
   - Review cache hit rates
   - Adjust TTL and cache size
3. **Optimize Circuit Breaker**:
   - Review failure patterns
   - Adjust thresholds

### Deployment Procedures

#### Pre-deployment Checklist
- [ ] All tests passing
- [ ] Security scan completed
- [ ] Performance tests completed
- [ ] Configuration reviewed
- [ ] Rollback plan prepared

#### Deployment Steps
1. **Deploy to Staging**:
   ```bash
   helm upgrade stock-service-staging ./charts/stock-service \
     --namespace staging \
     --set image.tag=new-version
   ```
2. **Run Smoke Tests**:
   ```bash
   curl http://staging-endpoint/health
   curl http://staging-endpoint/metrics
   ```
3. **Monitor for 30 minutes**:
   - Check error rates
   - Verify response times
   - Monitor resource usage
4. **Deploy to Production**:
   ```bash
   helm upgrade stock-service ./charts/stock-service \
     --namespace stock-service \
     --set image.tag=new-version
   ```
5. **Verify Deployment**:
   - Check pod status
   - Verify service endpoints
   - Monitor key metrics

#### Post-deployment Verification
1. **Health Check**:
   ```bash
   curl https://your-domain/health
   ```
2. **Metrics Verification**:
   ```bash
   curl https://your-domain/metrics | grep stock_service
   ```
3. **Traffic Validation**:
   - Monitor for 15 minutes
   - Verify error rates < 1%
   - Check response times

## Disaster Recovery

### Backup Strategy

#### Configuration Backup
```bash
# Backup Helm values
helm get values stock-service -n stock-service > stock-service-values-backup.yaml

# Backup Kubernetes manifests
kubectl get all -n stock-service -o yaml > stock-service-manifests-backup.yaml
```

#### Database Backup (if applicable)
- No persistent database in current architecture
- Cache is ephemeral and will repopulate

### Recovery Procedures

#### Complete Service Failure
1. **Check Cluster Status**:
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces
   ```
2. **Restore from Backup**:
   ```bash
   # Restore configuration
   helm upgrade stock-service ./charts/stock-service \
     --namespace stock-service \
     -f stock-service-values-backup.yaml
   ```
3. **Verify Recovery**:
   ```bash
   kubectl get pods -n stock-service
   curl https://your-domain/health
   ```

#### Data Center Failure
1. **Switch to DR Cluster**:
   - Update DNS to point to DR cluster
   - Verify service availability
2. **Restore Configuration**:
   - Apply backed-up manifests
   - Verify all settings
3. **Monitor Recovery**:
   - Check service health
   - Verify data consistency

### Business Continuity

#### RTO (Recovery Time Objective): 15 minutes
#### RPO (Recovery Point Objective): 5 minutes (configuration only)

#### Communication Plan
1. **Internal**: Notify via Slack #incidents
2. **External**: Update status page
3. **Stakeholders**: Send recovery notification

## Contact Information

### On-Call Rotation
- **Primary**: SRE Team Lead
- **Secondary**: Platform Engineering Team
- **Escalation**: Engineering Manager

### Communication Channels
- **Slack**: #alerts, #incidents
- **Email**: sre-team@company.com
- **Phone**: +1-XXX-XXX-XXXX (Emergency only)

### External Dependencies
- **Polygon.io Support**: support@polygon.io
- **Cloud Provider**: AWS Support (if applicable)
- **CDN Provider**: CloudFlare Support (if applicable)

---

**Document Version**: 1.0
**Last Updated**: $(date +%Y-%m-%d)
**Next Review**: $(date -d "+3 months" +%Y-%m-%d)