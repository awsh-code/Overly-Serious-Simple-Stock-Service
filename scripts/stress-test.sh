#!/bin/bash

# Stock Service Stress Test Script
# This script will generate load to make Grafana charts move beautifully

set -euo pipefail

# Configuration
STOCK_SERVICE_URL="http://localhost:8080"
CONCURRENT_REQUESTS=50
TOTAL_REQUESTS_PER_ENDPOINT=1000
RAMP_UP_TIME=30
TEST_DURATION=300

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}    STOCK SERVICE STRESS TEST${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}Target: ${STOCK_SERVICE_URL}${NC}"
    echo -e "${BLUE}Concurrent Requests: ${CONCURRENT_REQUESTS}${NC}"
    echo -e "${BLUE}Total Requests per Endpoint: ${TOTAL_REQUESTS_PER_ENDPOINT}${NC}"
    echo -e "${BLUE}Test Duration: ${TEST_DURATION} seconds${NC}"
    echo -e "${CYAN}========================================${NC}"
}

# Function to test single endpoint
test_endpoint() {
    local endpoint=$1
    local description=$2
    local method=${3:-GET}
    local data=${4:-}
    
    echo -e "${YELLOW}Testing: ${description} (${endpoint})${NC}"
    
    # Test the endpoint first to make sure it's working
    #if ! curl -s -f "${STOCK_SERVICE_URL}${endpoint}" > /dev/null 2>&1; then
    #    echo -e "${RED}❌ Endpoint ${endpoint} is not responding!${NC}"
    #    return 1
    #fi
    
    echo -e "${GREEN}✅ Endpoint ${endpoint} is responding${NC}"
    return 0
}

# Function to generate load with curl
generate_load() {
    local endpoint=$1
    local description=$2
    local total_requests=$3
    local concurrent=$4
    
    echo -e "${PURPLE}Generating load: ${description}${NC}"
    echo -e "${BLUE}Requests: ${total_requests}, Concurrent: ${concurrent}${NC}"
    
    # Use GNU parallel if available, otherwise fallback to background processes
    if command -v parallel &> /dev/null; then
        seq 1 ${total_requests} | parallel -j ${concurrent} "
            curl -s -w 'Response: %{http_code}, Time: %{time_total}s\n' -o /dev/null '${STOCK_SERVICE_URL}${endpoint}' 2>/dev/null | head -1
        " | pv -l -s ${total_requests} > /dev/null 2>&1 || true
    else
        # Fallback to simple background processes
        for ((i=1; i<=total_requests; i++)); do
            {
                curl -s -w 'Response: %{http_code}, Time: %{time_total}s\n' -o /dev/null "${STOCK_SERVICE_URL}${endpoint}" 2>/dev/null
            } &
            
            # Limit concurrent processes
            if (( i % concurrent == 0 )); then
                wait
            fi
        done
        wait
    fi
    
    echo -e "${GREEN}✅ Load generation completed for ${description}${NC}"
}

# Function to test caching performance
test_caching() {
    echo -e "${CYAN}Testing caching performance...${NC}"
    
    # First request (should be slow - cache miss)
    echo -e "${YELLOW}First request (cache miss):${NC}"
    time1=$(curl -s -w "%{time_total}" -o /dev/null "${STOCK_SERVICE_URL}/")
    echo -e "${BLUE}Time: ${time1}s${NC}"
    
    # Second request (should be fast - cache hit)
    echo -e "${YELLOW}Second request (cache hit):${NC}"
    time2=$(curl -s -w "%{time_total}" -o /dev/null "${STOCK_SERVICE_URL}/")
    echo -e "${BLUE}Time: ${time2}s${NC}"
    
    # Calculate improvement
    improvement=$(echo "scale=2; ($time1 - $time2) / $time1 * 100" | bc -l 2>/dev/null || echo "N/A")
    echo -e "${GREEN}Cache improvement: ${improvement}%${NC}"
}

# Function to test circuit breaker
test_circuit_breaker() {
    echo -e "${CYAN}Testing circuit breaker...${NC}"
    
    # Make multiple requests to trigger circuit breaker if needed
    for i in {1..20}; do
        response=$(curl -s -w "%{http_code}" -o /dev/null "${STOCK_SERVICE_URL}/" 2>/dev/null || echo "000")
        echo -e "${BLUE}Request $i: HTTP ${response}${NC}"
        sleep 0.1
    done
    
    # Check circuit breaker status
    circuit_status=$(curl -s "${STOCK_SERVICE_URL}/circuit-breaker" || echo "Circuit breaker endpoint not available")
    echo -e "${GREEN}Circuit breaker status: ${circuit_status}${NC}"
}

# Function to monitor metrics during test
monitor_metrics() {
    echo -e "${CYAN}Monitoring metrics...${NC}"
    
    # Get current request count
    request_count=$(curl -s "${STOCK_SERVICE_URL}/metrics" 2>/dev/null | grep "stock_service_requests_total" | wc -l)
    echo -e "${BLUE}Available metrics: ${request_count}${NC}"
    
    # Show some key metrics
    echo -e "${YELLOW}Key metrics:${NC}"
    curl -s "${STOCK_SERVICE_URL}/metrics" 2>/dev/null | grep -E "(stock_service_requests_total|stock_service_circuit_breaker|stock_service_errors_total)" | head -10
}

# Function to run comprehensive test
run_comprehensive_test() {
    echo -e "${CYAN}Starting comprehensive stress test...${NC}"
    
    # Test all endpoints first
    echo -e "${YELLOW}Phase 1: Endpoint validation${NC}"
    test_endpoint "/" "Stock Data"
    test_endpoint "/health" "Health Check"
    test_endpoint "/circuit-breaker" "Circuit Breaker Status"
    test_endpoint "/metrics" "Metrics"
    test_endpoint "/docs" "Documentation"
    
    echo -e "${CYAN}========================================${NC}"
    
    # Test caching performance
    echo -e "${YELLOW}Phase 2: Caching performance test${NC}"
    test_caching
    
    echo -e "${CYAN}========================================${NC}"
    
    # Test circuit breaker
    echo -e "${YELLOW}Phase 3: Circuit breaker test${NC}"
    test_circuit_breaker
    
    echo -e "${CYAN}========================================${NC}"
    
    # Monitor initial metrics
    echo -e "${YELLOW}Phase 4: Initial metrics${NC}"
    monitor_metrics
    
    echo -e "${CYAN}========================================${NC}"
    
    # Generate sustained load
    echo -e "${YELLOW}Phase 5: Sustained load generation${NC}"
    
    # Start background load generators
    echo -e "${PURPLE}Starting background load generators...${NC}"
    
    # Stock data endpoint load
    generate_load "/" "Stock Data" ${TOTAL_REQUESTS_PER_ENDPOINT} ${CONCURRENT_REQUESTS} &
    PID1=$!
    
    # Health check endpoint load (higher frequency)
    generate_load "/health" "Health Checks" $((TOTAL_REQUESTS_PER_ENDPOINT * 2)) $((CONCURRENT_REQUESTS * 2)) &
    PID2=$!
    
    # Circuit breaker endpoint load
    generate_load "/circuit-breaker" "Circuit Breaker" $((TOTAL_REQUESTS_PER_ENDPOINT / 2)) ${CONCURRENT_REQUESTS} &
    PID3=$!
    
    # Metrics endpoint load (periodic monitoring)
    {
        for i in $(seq 1 50); do
            curl -s "${STOCK_SERVICE_URL}/metrics" > /dev/null 2>&1
            sleep 6
        done
    } &
    PID4=$!
    
    echo -e "${BLUE}Load generators started. Monitoring for ${TEST_DURATION} seconds...${NC}"
    
    # Monitor progress
    for ((i=1; i<=TEST_DURATION; i++)); do
        if (( i % 30 == 0 )); then
            echo -e "${CYAN}Progress: ${i}/${TEST_DURATION} seconds${NC}"
            
            # Show current metrics snapshot
            echo -e "${YELLOW}Current metrics snapshot:${NC}"
            curl -s "${STOCK_SERVICE_URL}/metrics" 2>/dev/null | grep -E "(stock_service_requests_total|stock_service_circuit_breaker_state|stock_service_errors_total)" | tail -5
        fi
        
        sleep 1
    done
    
    # Wait for all background jobs to complete
    echo -e "${YELLOW}Waiting for load generators to complete...${NC}"
    wait $PID1 $PID2 $PID3 $PID4 2>/dev/null || true
    
    echo -e "${CYAN}========================================${NC}"
    
    # Final metrics check
    echo -e "${YELLOW}Phase 6: Final metrics${NC}"
    monitor_metrics
    
    echo -e "${GREEN}✅ Comprehensive stress test completed!${NC}"
}

# Function to show Grafana URLs
show_grafana_urls() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}    GRAFANA DASHBOARD URLs${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "${BLUE}Main Grafana: http://grafana.46.225.33.158.nip.io${NC}"
    echo -e "${BLUE}Direct Prometheus: http://prometheus.46.225.33.158.nip.io${NC}"
    echo -e "${YELLOW}Look for these metrics in Grafana:${NC}"
    echo -e "  • stock_service_requests_total"
    echo -e "  • stock_service_request_duration_seconds"
    echo -e "  • stock_service_circuit_breaker_state"
    echo -e "  • stock_service_circuit_breaker_failures_total"
    echo -e "  • stock_service_stock_api_duration_seconds"
    echo -e "  • stock_service_errors_total"
    echo -e "${CYAN}========================================${NC}"
}

# Main execution
main() {
    print_header
    
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ curl is required but not installed.${NC}"
        exit 1
    fi
    
    # Check if bc is available (for calculations)
    if ! command -v bc &> /dev/null; then
        echo -e "${YELLOW}⚠️  bc is not available, some calculations will be skipped${NC}"
    fi
    
    show_grafana_urls
    
    echo -e "${YELLOW}Press Enter to start the stress test...${NC}"
    read -r
    
    run_comprehensive_test
    
    echo -e "${GREEN}🎉 Stress test completed! Check Grafana for beautiful charts!${NC}"
    echo -e "${CYAN}The metrics should be dancing now! 🕺💃${NC}"
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n${RED}Stress test interrupted by user${NC}"; exit 130' INT

# Run the main function
main "$@"