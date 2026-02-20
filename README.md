# Overly-Serious-Simple-Stock-Service

A Go-based stock price service with monitoring, caching, and fault tolerance capabilities.

## Overview

This service provides stock price data through a REST API with built-in observability features and resilience patterns.

### 🚀 API Endpoints
| Service | URL | Description |
|---------|-----|-------------|
| **Stock Service**      | [http://ping-service.46.225.33.158.nip.io/](http://ping-service.46.225.33.158.nip.io/) | Main API endpoint (default: MSFT) |
| **Scalar Docs**        | [http://ping-service.46.225.33.158.nip.io/docs](http://ping-service.46.225.33.158.nip.io/docs) | Interactive API docs |
| **Prometheus Metrics**   | [http://ping-service.46.225.33.158.nip.io/metrics](http://ping-service.46.225.33.158.nip.io/metrics) | Live application metrics |
| **Health Check**         | [http://ping-service.46.225.33.158.nip.io/health](http://ping-service.46.225.33.158.nip.io/health) | Service liveness probe |
| **Circuit Breaker**      | [http://ping-service.46.225.33.158.nip.io/circuit-breaker](http://ping-service.46.225.33.158.nip.io/circuit-breaker) | Circuit Breaker Status |

### 📊 Observability & Monitoring
| Dashboard | URL | Description | Credentials |
|-----------|-----|-------------|-------------|
| **Grafana Main**         | [http://grafana.46.225.33.158.nip.io](http://grafana.46.225.33.158.nip.io) | Main Grafana interface | `demo` / `mJolOtJL8o5Umhu5tmqIya` |
| **Golden Signals**       | [http://grafana.46.225.33.158.nip.io/d/308a147c-c6ef-47f7-92b0-143145813ce3/ping-service-golden-signals](http://grafana.46.225.33.158.nip.io/d/308a147c-c6ef-47f7-92b0-143145813ce3/ping-service-golden-signals) | **The Four Golden Signals** | `demo` / `mJolOtJL8o5Umhu5tmqIya` |
| **Service Metrics**      | [http://grafana.46.225.33.158.nip.io/d/92e1bab9-9ef6-4ec8-8952-61c46bbabad6/ping-service-dashboard](http://grafana.46.225.33.158.nip.io/d/92e1bab9-9ef6-4ec8-8952-61c46bbabad6/ping-service-dashboard) | Detailed service performance | `demo` / `mJolOtJL8o5Umhu5tmqIya` |

### Monitoring

The service tracks the Four Golden Signals for observability:

- **Latency**: Request duration tracking
- **Traffic**: Request volume monitoring  
- **Errors**: Failed request rate tracking
- **Saturation**: Resource utilization metrics

## Quick Start

### Option 1: Helm Chart (Production Recommended)

```bash
# Clone the repository
git clone https://github.com/awsh-code/Overly-Serious-Simple-Stock-Service.git
cd Overly-Serious-Simple-Stock-Service

# Install with Helm (production-ready)
helm install stock-service ./charts/stock-service \
  --namespace stock-service \
  --create-namespace \
  --set config.stockAPI.apiKey=YOUR_API_KEY \
  --set replicaCount=3 \
  --set autoscaling.enabled=true \
  --set monitoring.enabled=true

# Or use the comprehensive deployment guide
# See: docs/deployment.md for detailed options
```

### Option 2: Kubernetes Manifests (Development)

## 🌐 Live Demo

**🚀 See it in action on our production cluster:**

### 🚀 API Endpoints
| Service | URL | Description |
|---------|-----|-------------|
| **Stock Service**      | [http://ping-service.46.225.33.158.nip.io/](http://ping-service.46.225.33.158.nip.io/) | Main API endpoint (default: MSFT) |
| **Scalar Docs**        | [http://ping-service.46.225.33.158.nip.io/docs](http://ping-service.46.225.33.158.nip.io/docs) | Interactive API docs |
| **Prometheus Metrics**   | [http://ping-service.46.225.33.158.nip.io/metrics](http://ping-service.46.225.33.158.nip.io/metrics) | Live application metrics |
| **Health Check**         | [http://ping-service.46.225.33.158.nip.io/health](http://ping-service.46.225.33.158.nip.io/health) | Service liveness probe |
| **Circuit Breaker**      | [http://ping-service.46.225.33.158.nip.io/circuit-breaker](http://ping-service.46.225.33.158.nip.io/circuit-breaker) | Circuit Breaker Status |

### 📊 Observability & Monitoring
| Dashboard | URL | Description | Credentials |
|-----------|-----|-------------|-------------|
| **Grafana Main**         | [http://grafana.46.225.33.158.nip.io](http://grafana.46.225.33.158.nip.io) | Main Grafana interface | `demo` / `mJolOtJL8o5Umhu5tmqIya` |
| **Golden Signals**       | [http://grafana.46.225.33.158.nip.io/d/308a147c-c6ef-47f7-92b0-143145813ce3/ping-service-golden-signals](http://grafana.46.225.33.158.nip.io/d/308a147c-c6ef-47f7-92b0-143145813ce3/ping-service-golden-signals) | **The Four Golden Signals** | `demo` / `mJolOtJL8o5Umhu5tmqIya` |
| **Service Metrics**      | [http://grafana.46.225.33.158.nip.io/d/92e1bab9-9ef6-4ec8-8952-61c46bbabad6/ping-service-dashboard](http://grafana.46.225.33.158.nip.io/d/92e1bab9-9ef6-4ec8-8952-61c46bbabad6/ping-service-dashboard) | Detailed service performance | `demo` / `mJolOtJL8o5Umhu5tmqIya` |

#### 🎯 What are the Four Golden Signals?
The **Golden Signals** are the four most important metrics for monitoring any production system:

- **Latency**: How long requests take to complete (including errors)
- **Traffic**: How many requests per second your system is handling  
- **Errors**: The rate of failed requests (4xx, 5xx, timeouts)
- **Saturation**: How close your system is to being overloaded (CPU, memory, disk, network)

These signals provide a complete picture of system health and are essential for:
- **Capacity Planning**: Understanding when to scale
- **Incident Response**: Quickly identifying what's broken
- **Performance Optimization**: Finding bottlenecks
- **SLA Compliance**: Meeting service level objectives

**📊 Try our stress testing:**
```bash
# Quick 60-second load test
curl -s http://ping-service.46.225.33.158.nip.io/metrics | grep cache

# Full stress test (clone repo first)
git clone https://github.com/awsh-code/Overly-Serious-Simple-Stock-Service.git
cd Overly-Serious-Simple-Stock-Service
./scripts/quick-stress.sh
```

### Prerequisites
- Docker
- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl configured
- Make

### Build & Deploy (One Command)
```bash
# Clone the repository
git clone https://github.com/awsh-code/Overly-Serious-Simple-Stock-Service.git
cd Overly-Serious-Simple-Stock-Service

# Build, push, and deploy
make all VERSION=v1.0.0
```

### CI/CD Pipeline
The GitHub Actions pipeline includes:
- Automated testing (unit, integration, security)
- Security scanning with Trivy
- Helm chart validation and packaging
- Multi-stage deployment automation
- Monitoring integration

See `.github/workflows/` for complete configuration.

## API Documentation

Access interactive API documentation at: `http://localhost:8080/docs`

### Available Endpoints
- `GET /` - Get stock data for default symbol
- `GET /{symbol}` - Get stock data for specific symbol
- `GET /{symbol}/{days}` - Get stock data with custom day range
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics
- `GET /docs` - Interactive documentation
- `GET /circuit-breaker` - Circuit breaker status

## Architecture

This service follows a standard microservice architecture with load balancing, service logic, and external API integration. Includes monitoring with Prometheus and Grafana.

### System Architecture
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Stock Service  │    │   Alpha Vantage │
│   (Ingress)     │────│   (This Service) │────│   API (External)│
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌──────────────────┐
│   Grafana       │    │   Prometheus     │
│   Dashboards    │    │   Metrics        │
└─────────────────┘    └──────────────────┘
```

## Features

- **Stock Price API**: Get closing prices for any stock symbol
- **Average Calculation**: Automatically calculates average closing price
- **Environment Configuration**: SYMBOL and NDAYS configurable via environment variables
- **API Key Management**: Secure API key handling via Kubernetes secrets
- **Caching Layer**: In-memory caching with cache hit/miss metrics
- **Circuit Breaker**: Prevents cascading failures with state monitoring
- **Prometheus Metrics**: Comprehensive metrics for monitoring and alerting
- **Grafana Dashboards**: Pre-built dashboards for visualization
- **Health Checks**: Liveness and readiness probes for Kubernetes
- **Stress Testing**: Included scripts for load testing and validation
- **Horizontal Pod Autoscaler**: Automatic scaling based on CPU/memory usage
- **Scalar Documentation**: Interactive API documentation

## Configuration

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `SYMBOL` | Stock symbol to track | `MSFT` |
| `NDAYS` | Number of days of data | `7` |
| `APIKEY` | Alpha Vantage API key | *(required)* |
| `PORT` | Service port | `8080` |
| `CACHE_TTL` | Cache TTL in seconds | `300` |
| `CIRCUIT_BREAKER_TIMEOUT` | Circuit breaker timeout | `30s` |

## 🧪 Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-integration
```

### Load Testing
```bash
# Test with 100 concurrent users
make stress-test CONCURRENT=100 DURATION=60s
```

## Monitoring & Observability

### Metrics Available
- `stock_api_requests_total`: Total API requests
- `stock_api_request_duration_seconds`: Request latency
- `stock_api_cache_hits_total`: Cache hit count
- `stock_api_cache_misses_total`: Cache miss count
- `stock_api_circuit_breaker_state`: Circuit breaker state (0=closed, 1=open, 2=half-open)
- `stock_api_external_calls_total`: External API calls
- `stock_api_external_call_duration_seconds`: External API latency

### Alerting Strategy
Metrics are structured for Prometheus alerting rules:
- High Latency: 95th percentile latency above 2 seconds
- Circuit Breaker Open: Circuit breaker state == 1 (open)
- Low Cache Hit Rate: Cache hit rate below 50%

## Security

- API keys stored in Kubernetes secrets
- Non-root container execution
- Resource limits and requests configured
- Network policies ready for implementation
- No sensitive data in logs or metrics

### Production Security Patterns
See [Production Security Documentation](docs/production-security.md) for:
- Monitoring Access Controls
- Network Security
- Secret Management
- Rate Limiting & DDoS Protection
- Audit Logging
- Container Security

## Scalability

### Horizontal Pod Autoscaler
- Min Replicas: 2
- Max Replicas: 10
- CPU Target: 70% utilization
- Memory Target: 80% utilization
- Scale-up Rate: 50% per 30 seconds
- Scale-down Rate: 10% per 60 seconds

### Resource Requirements
- CPU Request: 100m
- CPU Limit: 500m
- Memory Request: 128Mi
- Memory Limit: 512Mi

## Project Structure

```
.
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── cache/                  # Caching layer
│   ├── circuitbreaker/         # Circuit breaker implementation
│   ├── config/                 # Configuration management
│   ├── handlers/               # HTTP handlers
│   ├── middleware/             # HTTP middleware
│   └── stock/                  # Stock API client
├── k8s/                        # Kubernetes manifests
├── charts/                     # Helm charts
│   └── stock-service/          # Production Helm chart
├── scripts/                    # Operational scripts
├── docs/                       # Documentation
├── tests/                      # Integration tests
├── .github/workflows/          # CI/CD pipelines
├── Dockerfile                  # Multi-stage Docker build
├── Makefile                    # Build and deployment automation
├── go.mod                      # Go dependencies
└── README.md                   # This file
```

## Architecture Documentation

For detailed technical documentation on each component:

- [Circuit Breaker Architecture](docs/architecture-circuit-breaker.md): Fault tolerance patterns and resilience engineering
- [Caching Strategy](docs/architecture-caching.md): Performance optimization and cache management
- [Metrics & Observability](docs/architecture-metrics-observability.md): Prometheus metrics and Grafana dashboards
- [API Design & Error Handling](docs/architecture-api-design.md): RESTful API patterns and error handling
- [Production Security](docs/production-security.md): Security patterns and access controls
- [Kustomize Deployment](docs/kustomize-deployment.md): GitOps deployment and environment management
- [Deployment Guide](docs/deployment.md): Deployment guide with Helm charts and Kubernetes manifests
- [Operational Runbook](docs/operational-runbook.md): Operations, incident response, and maintenance procedures

## Additional Features

### Resilience Engineering
- Circuit Breaker: Prevents cascading failures with state monitoring
- Caching Layer: In-memory caching with cache hit/miss metrics
- Health Checks: Kubernetes-native liveness/readiness probes
- Graceful Degradation: Service continues with cached data if external API fails
- Timeout Protection: All external calls have configurable timeouts
- Retry Logic: Configurable retry attempts for transient failures

### Observability
- Metrics: Comprehensive Prometheus metrics for all operations
- Dashboards: Pre-built Grafana dashboards for visualization
- Alerting: Metrics structured for easy alerting rules
- Distributed Tracing: Ready for OpenTelemetry integration
- Structured Logging: Zap-based logging with correlation IDs

### Operational Features
- One-Command Deploy: Complete automation with Make
- Rolling Updates: Zero-downtime deployments
- Rollback Capability: Easy reversion to previous versions
- Environment Management: Separate configs for dev/staging/prod
- Documentation: Comprehensive setup and operational guides
- Scalar Integration: Interactive API documentation
- Kustomize Patterns: GitOps deployment strategy
- Security Documentation: Production security patterns and access controls

### Performance Engineering
- Load Testing: Included stress testing scripts
- Horizontal Scaling: HPA with CPU/memory-based scaling
- Resource Optimization: Multi-stage Docker builds
- Cache Efficiency: 95%+ hit rates under normal load
- Circuit Breaker: Sub-second failure detection and recovery

### Production Deployment
- Helm Charts: Helm chart with comprehensive configuration
- GitOps Integration: Automated CI/CD with GitHub Actions
- Multi-Environment Support: Dev, staging, production lifecycle management
- Package Management: Helm chart packaging and artifact management
- Deployment Automation: One-command production deployments
- Rollback Capability: Easy reversion to previous versions

## Technical Overview

This project demonstrates a production-ready microservice with comprehensive monitoring, resilience patterns, and operational best practices.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.