# Makefile for the Overly-Serious-Simple-Stock-Service
# Production-grade stock ticker microservice

# --- Variables ---
# Image configuration
IMAGE_NAME ?= codyadkinsdev/stock-service
VERSION ?= latest
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_HASH := $(VERSION)-$(GIT_HASH)

# Application details
APP_NAME = stock-service
NAMESPACE = default

# Kubernetes configuration
KUBECONFIG_PATH ?= $(HOME)/.kube/config
K8S_DIR = k8s

# Docker configuration
DOCKER_PLATFORM = linux/amd64,linux/arm64

# --- Targets ---
.PHONY: all build push deploy restart status logs test clean k8s-apply k8s-delete port-forward

# Default target: build, push, and deploy
all: build push deploy wait status
	@echo "✅ Full deployment cycle complete for $(APP_NAME):$(VERSION_HASH)"

# Build multi-platform Docker image
build:
	@echo "🏗️  Building multi-platform Docker image: $(IMAGE_NAME):$(VERSION_HASH)..."
	@docker buildx build --platform $(DOCKER_PLATFORM) -t $(IMAGE_NAME):$(VERSION_HASH) --push .
	@echo "✅ Docker image built and pushed successfully"

# Alternative build for local development (single platform)
build-local:
	@echo "🏗️  Building local Docker image: $(IMAGE_NAME):$(VERSION_HASH)..."
	@docker build -t $(IMAGE_NAME):$(VERSION_HASH) .
	@echo "✅ Local Docker image built successfully"

# Deploy to Kubernetes
deploy: k8s-apply
	@echo "🚢  Deployment initiated for $(APP_NAME):$(VERSION_HASH)"

# Apply Kubernetes manifests
k8s-apply:
	@echo "📋  Applying Kubernetes manifests..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f $(K8S_DIR)/
	@echo "✅ Kubernetes manifests applied"

# Delete Kubernetes resources
k8s-delete:
	@echo "🗑️  Deleting Kubernetes resources..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl delete -f $(K8S_DIR)/ --ignore-not-found=true
	@echo "✅ Kubernetes resources deleted"

# Restart deployment
restart:
	@echo "🔄  Restarting deployment $(APP_NAME)..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/$(APP_NAME) -n $(NAMESPACE)
	@echo "✅ Deployment restarted"

# Wait for deployment to be ready
wait:
	@echo "⏳  Waiting for deployment to be ready..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl wait --for=condition=available deployment/$(APP_NAME) -n $(NAMESPACE) --timeout=120s || echo "⚠️  Deployment wait timed out, checking status..."

# Get deployment status
status:
	@echo "ℹ️  Deployment Status:"
	@echo "   Pods:"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get pods -n $(NAMESPACE) -l app=$(APP_NAME)
	@echo "   Deployment:"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get deployment $(APP_NAME) -n $(NAMESPACE)
	@echo "   Service:"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get service $(APP_NAME) -n $(NAMESPACE)
	@echo "   HPA:"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get hpa $(APP_NAME)-hpa -n $(NAMESPACE) 2>/dev/null || echo "   HPA not found"

# Tail application logs
logs:
	@echo "📜  Tailing logs for $(APP_NAME)..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -f -n $(NAMESPACE) -l app=$(APP_NAME) --tail=50

# Get all logs from all pods
logs-all:
	@echo "📜  Getting all logs for $(APP_NAME)..."
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl logs -n $(NAMESPACE) -l app=$(APP_NAME) --all-containers=true --tail=100

# Port forward to access service locally
port-forward:
	@echo "🔌  Port forwarding $(APP_NAME) service to localhost:8080..."
	@echo "   Press Ctrl+C to stop"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl port-forward -n $(NAMESPACE) service/$(APP_NAME) 8080:80

# Port forward to Grafana
port-forward-grafana:
	@echo "📊  Port forwarding Grafana to localhost:3001..."
	@echo "   Press Ctrl+C to stop"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl port-forward -n monitoring service/prometheus-grafana 3001:80

# Port forward to Prometheus
port-forward-prometheus:
	@echo "📈  Port forwarding Prometheus to localhost:9090..."
	@echo "   Press Ctrl+C to stop"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl port-forward -n monitoring service/prometheus-kube-prometheus-prometheus 9090:9090

# Run stress test
stress-test:
	@echo "🔥  Running stress test..."
	@./scripts/stress-test.sh

# Run quick stress test
quick-stress:
	@echo "⚡  Running quick stress test..."
	@./scripts/quick-stress.sh

# Run unit tests
test:
	@echo "🧪  Running unit tests..."
	@go test -v ./...

# Run integration tests
test-integration:
	@echo "🔗  Running integration tests..."
	@go test -v -tags=integration ./tests/...

# Clean up resources
clean: k8s-delete
	@echo "🧹  Cleaning up Docker images..."
	@docker rmi $(IMAGE_NAME):$(VERSION_HASH) 2>/dev/null || true
	@echo "✅ Cleanup complete"

# Development setup
dev-setup:
	@echo "🔧  Setting up development environment..."
	@echo "   Installing dependencies..."
	@go mod download
	@echo "   Building binary..."
	@go build -o bin/stock-service ./cmd/main.go
	@echo "✅ Development setup complete"

# Show help
help:
	@echo "Overly-Serious-Simple-Stock-Service Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all              - Build, push, and deploy (default)"
	@echo "  build            - Build multi-platform Docker image"
	@echo "  build-local      - Build local Docker image (single platform)"
	@echo "  deploy           - Deploy to Kubernetes"
	@echo "  restart          - Restart deployment"
	@echo "  wait             - Wait for deployment to be ready"
	@echo "  status           - Show deployment status"
	@echo "  logs             - Tail application logs"
	@echo "  logs-all         - Get all logs from all pods"
	@echo "  port-forward     - Port forward service to localhost:8080"
	@echo "  port-forward-grafana - Port forward Grafana to localhost:3001"
	@echo "  port-forward-prometheus - Port forward Prometheus to localhost:9090"
	@echo "  stress-test      - Run stress test"
	@echo "  quick-stress     - Run quick stress test"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  clean            - Clean up all resources"
	@echo "  dev-setup        - Setup development environment"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE_NAME       - Docker image name (default: $(IMAGE_NAME))"
	@echo "  VERSION          - Version tag (default: $(VERSION))"
	@echo "  NAMESPACE        - Kubernetes namespace (default: $(NAMESPACE))"
	@echo "  KUBECONFIG_PATH  - Path to kubeconfig (default: $(KUBECONFIG_PATH))"

# Show version information
version:
	@echo "Stock Service Version Information:"
	@echo "  Image: $(IMAGE_NAME):$(VERSION_HASH)"
	@echo "  Git Hash: $(GIT_HASH)"
	@echo "  Namespace: $(NAMESPACE)"
	@echo "  Kubeconfig: $(KUBECONFIG_PATH)"