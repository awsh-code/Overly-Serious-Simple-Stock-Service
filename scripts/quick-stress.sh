#!/bin/bash

# Quick Stress Test - Immediate Grafana Action

echo "🚀 LAUNCHING IMMEDIATE GRAFANA ACTION! 🚀"
echo "Target: http://ping-service.46.225.33.158.nip.io"
echo "Duration: 60 seconds of pure metric chaos!"
echo ""
echo "Watch these metrics in Grafana:"
echo "  • ping_service_requests_total"
echo "  • ping_service_request_duration_seconds" 
echo "  • ping_service_circuit_breaker_state"
echo "  • ping_service_stock_api_duration_seconds"
echo ""
echo "Starting in 3... 2... 1... 🔥"

# Generate immediate load across all endpoints
for i in {1..60}; do
    # Stock data requests (with cache testing)
    curl -s "http://ping-service.46.225.33.158.nip.io/" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/" > /dev/null 2>&1 &
    
    # Health check spam (Kubernetes probes simulation)
    curl -s "http://ping-service.46.225.33.158.nip.io/health" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/health" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/health" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/health" > /dev/null 2>&1 &
    curl -s "http://ping-service.46.225.33.158.nip.io/health" > /dev/null 2>&1 &
    
    # Circuit breaker monitoring
    curl -s "http://ping-service.46.225.33.158.nip.io/circuit-breaker" > /dev/null 2>&1 &
    
    # Metrics scraping (like Prometheus would do)
    curl -s "http://ping-service.46.225.33.158.nip.io/metrics" > /dev/null 2>&1 &
    
    echo -n "."
    sleep 1
    
    # Limit concurrent requests
    if (( i % 10 == 0 )); then
        wait
        echo ""
        echo "💥 ${i}/60 seconds completed - Metrics should be dancing! 💥"
    fi
done

wait
echo ""
echo "🎉 METRIC CHAOS COMPLETE! 🎉"
echo "Check Grafana now - your charts should be MOVING! 🕺💃"
echo ""
echo "Quick metrics check:"
curl -s "http://ping-service.46.225.33.158.nip.io/metrics" | grep -E "(ping_service_requests_total|ping_service_request_duration_seconds)" | tail -3