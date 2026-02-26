graph TB
    subgraph "External Systems"
        AV["Alpha Vantage API<br/>External Stock Data Provider"]
        K8s["Kubernetes Cluster<br/>Container Orchestration"]
        MON["Monitoring Stack<br/>Prometheus + Grafana"]
    end

    subgraph "Service Architecture"
        subgraph "Entry Point"
            MAIN["main.go<br/>Application Bootstrap"]
        end

        subgraph "HTTP Layer"
            ROUTER["Gorilla Mux Router<br/>HTTP Routing"]
            MIDDLE["Middleware Layer<br/>Logging & Metrics"]
            HANDLERS["HTTP Handlers<br/>Request Processing"]
        end

        subgraph "Business Logic"
            STOCK_CLIENT["Stock Client<br/>External API Integration"]
            CACHE["In-Memory Cache<br/>TTL-based Caching"]
            CB["Circuit Breaker<br/>Fault Tolerance"]
        end

        subgraph "Configuration"
            CONFIG["Configuration<br/>Environment Variables"]
        end

        subgraph "Observability"
            METRICS["Prometheus Metrics<br/>Performance Tracking"]
            HEALTH["Health Checks<br/>Liveness & Readiness"]
        end
    end

    subgraph "Deployment & Operations"
        DOCKER["Docker Container<br/>Multi-stage Build"]
        HELM["Helm Charts<br/>Kubernetes Packaging"]
        K8S_MAN["K8s Manifests<br/>Direct Deployment"]
        CI_CD["GitHub Actions<br/>CI/CD Pipeline"]
    end

    subgraph "API Endpoints"
        EP1["GET /<br/>Default Stock Data"]
        EP2["GET /{symbol}<br/>Specific Symbol"]
        EP3["GET /{symbol}/{days}<br/>Custom Range"]
        EP4["GET /health<br/>Health Check"]
        EP5["GET /ready<br/>Readiness Check"]
        EP6["GET /metrics<br/>Prometheus Metrics"]
        EP7["GET /docs<br/>API Documentation"]
        EP8["GET /circuit-breaker<br/>CB Status"]
    end

    subgraph "Key Features"
        F1["Caching Layer<br/>Performance Optimization"]
        F2["Circuit Breaker<br/>Resilience Pattern"]
        F3["Metrics Collection<br/>Four Golden Signals"]
        F4["Auto-scaling<br/>HPA Configuration"]
        F5["Security<br/>Secrets Management"]
        F6["Documentation<br/>Scalar Integration"]
    end

    %% Main flow connections
    MAIN --> CONFIG
    MAIN --> ROUTER
    MAIN --> METRICS
    
    ROUTER --> MIDDLE
    MIDDLE --> HANDLERS
    HANDLERS --> STOCK_CLIENT
    HANDLERS --> HEALTH
    
    STOCK_CLIENT --> CACHE
    STOCK_CLIENT --> CB
    STOCK_CLIENT --> AV
    
    CACHE --> F1
    CB --> F2
    
    %% Endpoint mappings
    HANDLERS --> EP1
    HANDLERS --> EP2
    HANDLERS --> EP3
    HANDLERS --> EP4
    HANDLERS --> EP5
    HANDLERS --> EP6
    HANDLERS --> EP7
    HANDLERS --> EP8
    
    %% Deployment connections
    DOCKER --> K8s
    HELM --> K8s
    K8S_MAN --> K8s
    CI_CD --> DOCKER
    
    %% Monitoring connections
    METRICS --> MON
    HEALTH --> K8s
    F3 --> MON
    
    %% Feature connections
    F1 --> CACHE
    F2 --> CB
    F3 --> METRICS
    F4 --> K8s
    F5 --> CONFIG
    F6 --> EP7

    %% Styling
    classDef external fill:#f9f,stroke:#333,stroke-width:2px
    classDef service fill:#bbf,stroke:#333,stroke-width:2px
    classDef deployment fill:#bfb,stroke:#333,stroke-width:2px
    classDef api fill:#fbf,stroke:#333,stroke-width:2px
    classDef feature fill:#fbb,stroke:#333,stroke-width:2px
    
    class AV,K8s,MON external
    class MAIN,ROUTER,MIDDLE,HANDLERS,STOCK_CLIENT,CACHE,CB,CONFIG,METRICS,HEALTH service
    class DOCKER,HELM,K8S_MAN,CI_CD deployment
    class EP1,EP2,EP3,EP4,EP5,EP6,EP7,EP8 api
    class F1,F2,F3,F4,F5,F6 feature
