# 🎯 Kustomize Deployment Strategy

This document explains our Kustomize-based deployment approach, demonstrating production-ready configuration management for Kubernetes applications.

## 🚀 Why Kustomize Over Helm?

While Helm is excellent for package management, **Kustomize excels at configuration management** without templating complexity:

- **No Templating**: Pure YAML, no Go templating syntax
- **Composable**: Layer configurations without complex logic
- **GitOps Friendly**: Native Kubernetes resource generation
- **Namespace Aware**: Easy multi-environment management

## 📁 Directory Structure

```
k8s/
├── base/                    # Base configurations
│   ├── kustomization.yaml   # Base resource definitions
│   ├── deployment.yaml      # Application deployment
│   ├── service.yaml         # Service definition
│   ├── configmap.yaml       # Non-sensitive configuration
│   └── secret.yaml          # Secret template (empty for GitOps)
├── overlays/
│   ├── development/         # Dev environment
│   │   ├── kustomization.yaml
│   │   ├── deployment-patch.yaml
│   │   └── configmap-patch.yaml
│   ├── staging/            # Staging environment
│   │   ├── kustomization.yaml
│   │   ├── deployment-patch.yaml
│   │   └── ingress-patch.yaml
│   └── production/         # Production environment
│       ├── kustomization.yaml
│       ├── deployment-patch.yaml
│       ├── ingress-patch.yaml
│       └── hpa-patch.yaml
└── components/
    ├── monitoring/         # Reusable monitoring stack
    │   ├── kustomization.yaml
    │   ├── prometheus.yaml
│   └── grafana.yaml
    └── security/           # Security policies
        ├── kustomization.yaml
        ├── network-policy.yaml
        └── pod-security-policy.yaml
```

## 🎯 Base Configuration (GitOps Ready)

### `k8s/base/kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

metadata:
  name: stock-service-base

resources:
- deployment.yaml
- service.yaml
- configmap.yaml
- secret.yaml

commonLabels:
  app.kubernetes.io/name: stock-service
  app.kubernetes.io/component: api
  app.kubernetes.io/managed-by: kustomize

images:
- name: codyadkinsdev/stock-service
  newTag: latest

configMapGenerator:
- name: stock-service-config
  literals:
  - SYMBOL=MSFT
  - NDAYS=7
  - CACHE_TTL=300
  - CIRCUIT_BREAKER_TIMEOUT=30s

secretGenerator:
- name: stock-service-secret
  literals:
  - apikey=PLACEHOLDER_API_KEY  # Replace in overlays
  options:
    disableNameSuffixHash: true
```

## 🚀 Environment Overlays

### Development Overlay
```yaml
# k8s/overlays/development/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: stock-service-dev

bases:
- ../../base

patchesStrategicMerge:
- deployment-patch.yaml

images:
- name: codyadkinsdev/stock-service
  newTag: dev-latest

replicas:
- name: stock-service
  count: 1

configMapGenerator:
- name: stock-service-config
  behavior: merge
  literals:
  - CACHE_TTL=60  # Shorter cache for dev
```

### Production Overlay
```yaml
# k8s/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: stock-service-prod

bases:
- ../../base
- ../../components/monitoring  # Include monitoring
- ../../components/security    # Include security policies

patchesStrategicMerge:
- deployment-patch.yaml
- ingress-patch.yaml
- hpa-patch.yaml

images:
- name: codyadkinsdev/stock-service
  newTag: v1.2.3  # Specific version

replicas:
- name: stock-service
  count: 3

configMapGenerator:
- name: stock-service-config
  behavior: merge
  literals:
  - CACHE_TTL=900  # Longer cache for prod
  - CIRCUIT_BREAKER_TIMEOUT=60s
```

## 🔧 Deployment Commands

### Development Deployment
```bash
# Deploy to development
kubectl apply -k k8s/overlays/development/

# Update with new image
kubectl apply -k k8s/overlays/development/ --dry-run=client
```

### Production Deployment
```bash
# Deploy to production (with approval)
kubectl apply -k k8s/overlays/production/ --dry-run=server
kubectl apply -k k8s/overlays/production/

# Verify deployment
kubectl rollout status deployment/stock-service -n stock-service-prod
```

### GitOps Integration (ArgoCD)
```yaml
# argocd-application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: stock-service
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/stock-service
    targetRevision: main
    path: k8s/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: stock-service-prod
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## 🎯 Advanced Kustomize Features

### 1. Component-Based Architecture
```yaml
# k8s/components/monitoring/kustomization.yaml
resources:
- prometheus-service-monitor.yaml
- grafana-dashboard.yaml
- alerting-rules.yaml

commonLabels:
  monitoring: enabled
```

### 2. Secret Management with SOPS
```yaml
# k8s/overlays/production/secret-generator.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

secretGenerator:
- name: stock-service-secret
  files:
  - apikey=secrets/apikey.encrypted  # Encrypted with SOPS
```

### 3. Image Policy with Kustomize
```yaml
# k8s/overlays/production/image-policy.yaml
images:
- name: codyadkinsdev/stock-service
  digest: sha256:abc123...  # Pin to specific digest
```

## 🚀 Benefits for Staff SRE Role

### 1. **Configuration as Code**
- All changes tracked in Git
- Rollback capability via Git history
- Peer review process for infrastructure changes

### 2. **Environment Consistency**
- Same base configuration across all environments
- Clear differences between dev/staging/prod
- No "it works on my machine" issues

### 3. **Security by Design**
- Secrets managed per environment
- Network policies applied consistently
- Security scanning integrated into pipeline

### 4. **Operational Excellence**
- Easy rollbacks with `kubectl apply -k`
- Clear resource ownership and labeling
- Integrated monitoring and alerting

## 🎯 Interview Talking Points

**When discussing Kustomize vs Helm**:
- "Kustomize provides **declarative configuration management** without the complexity of templating"
- "Perfect for **GitOps workflows** where we want to see the exact resources being deployed"
- "Allows us to **compose configurations** from reusable components while maintaining clarity"

**Security Benefits**:
- "Secrets are **environment-specific** and never committed to the base configuration"
- "Network policies and security contexts are **applied consistently** across all environments"
- "Image digests are **pinned to specific versions** to prevent supply chain attacks"

This Kustomize approach demonstrates **production-grade configuration management** that scales from a single service to an entire platform.