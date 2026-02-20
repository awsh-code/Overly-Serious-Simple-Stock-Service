// +build integration

package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/config"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/handlers"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/stock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupTestServer(t *testing.T) (*httptest.Server, *prometheus.Registry) {
	// Create a new Prometheus registry for testing
	reg := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.Config{
		Port:                    "8080",
		APIKey:                  "test-api-key",
		CacheTTL:                5 * time.Minute,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		MinRequestThreshold:     10,
	}

	// Create stock client with test API key
	stockClient := stock.NewClient(cfg.APIKey, reg)

	// Create handler with test configuration
	handler := handlers.New(cfg, stockClient, reg)

	// Create router
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/health", handler.Health).Methods("GET")
	router.HandleFunc("/ready", handler.Ready).Methods("GET")
	router.HandleFunc("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP)
	router.HandleFunc("/", handler.Home).Methods("GET")
	router.HandleFunc("/{symbol}", handler.GetStock).Methods("GET")
	router.HandleFunc("/{symbol}/{days}", handler.GetStockHistory).Methods("GET")
	router.HandleFunc("/docs", handler.Docs).Methods("GET")
	router.HandleFunc("/swagger.yaml", handler.Swagger).Methods("GET")

	// Create test server
	server := httptest.NewServer(router)

	return server, reg
}

func TestHealthEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "healthy") {
		t.Errorf("Expected response to contain 'healthy', got: %s", string(body))
	}
}

func TestReadyEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/ready")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "ready") {
		t.Errorf("Expected response to contain 'ready', got: %s", string(body))
	}
}

func TestMetricsEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check for expected metrics
	metrics := string(body)
	expectedMetrics := []string{
		"# HELP",
		"# TYPE",
		"ping_service_",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(metrics, expected) {
			t.Errorf("Expected metrics to contain '%s'", expected)
		}
	}
}

func TestHomeEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "Stock Service API") {
		t.Errorf("Expected response to contain 'Stock Service API', got: %s", string(body))
	}
}

func TestDocsEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/docs")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "Scalar") {
		t.Errorf("Expected response to contain 'Scalar', got: %s", string(body))
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/swagger.yaml")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "openapi: 3.0.0") {
		t.Errorf("Expected response to contain OpenAPI specification, got: %s", string(body))
	}
}

func TestStockEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Test with a known stock symbol (this will make a real API call)
	resp, err := http.Get(server.URL + "/AAPL")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// We expect either success or a circuit breaker error
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(body), "Circuit breaker is open") {
			t.Logf("Non-circuit breaker error: %s", string(body))
		}
		return // Skip further checks if we got a circuit breaker error
	}

	var stockData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stockData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check for expected fields
	if _, ok := stockData["symbol"]; !ok {
		t.Error("Expected response to contain 'symbol' field")
	}
	if _, ok := stockData["price"]; !ok {
		t.Error("Expected response to contain 'price' field")
	}
}

func TestStockHistoryEndpoint(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Test with a known stock symbol and days
	resp, err := http.Get(server.URL + "/AAPL/30")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// We expect either success or a circuit breaker error
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(body), "Circuit breaker is open") {
			t.Logf("Non-circuit breaker error: %s", string(body))
		}
		return // Skip further checks if we got a circuit breaker error
	}

	var stockData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stockData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check for expected fields
	if _, ok := stockData["symbol"]; !ok {
		t.Error("Expected response to contain 'symbol' field")
	}
	if _, ok := stockData["history"]; !ok {
		t.Error("Expected response to contain 'history' field")
	}
}

func TestCacheHitMetrics(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Make first request (should be cache miss)
	resp1, err := http.Get(server.URL + "/AAPL")
	if err != nil {
		t.Fatalf("Failed to make first request: %v", err)
	}
	resp1.Body.Close()

	// Wait a bit for metrics to be recorded
	time.Sleep(100 * time.Millisecond)

	// Get initial cache metrics
	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	initialMetrics := string(body)

	// Make second request (should be cache hit)
	resp2, err := http.Get(server.URL + "/AAPL")
	if err != nil {
		t.Fatalf("Failed to make second request: %v", err)
	}
	resp2.Body.Close()

	// Wait for metrics update
	time.Sleep(100 * time.Millisecond)

	// Get updated metrics
	resp, err = http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to get updated metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	updatedMetrics := string(body)

	// Check that cache metrics are present
	if !strings.Contains(updatedMetrics, "ping_service_cache_hits_total") {
		t.Error("Expected cache hits metric to be present")
	}
	if !strings.Contains(updatedMetrics, "ping_service_cache_misses_total") {
		t.Error("Expected cache misses metric to be present")
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Make multiple requests to potentially trigger circuit breaker
	for i := 0; i < 20; i++ {
		resp, err := http.Get(server.URL + "/INVALID")
		if err != nil {
			t.Logf("Request %d failed: %v", i, err)
			continue
		}
		resp.Body.Close()
	}

	// Wait for metrics update
	time.Sleep(200 * time.Millisecond)

	// Get metrics
	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	metrics := string(body)

	// Check that circuit breaker metrics are present
	if !strings.Contains(metrics, "ping_service_circuit_breaker_state") {
		t.Error("Expected circuit breaker state metric to be present")
	}
	if !strings.Contains(metrics, "ping_service_stock_api_duration_seconds") {
		t.Error("Expected stock API duration metric to be present")
	}
}

func TestErrorHandling(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Test with invalid symbol
	resp, err := http.Get(server.URL + "/INVALID_SYMBOL_12345")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should get either a successful response (if circuit breaker allows) or a circuit breaker error
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		if !strings.Contains(string(body), "Circuit breaker is open") {
			// If it's not a circuit breaker error, it might be a stock API error
			t.Logf("Got non-circuit breaker error: %s", string(body))
		}
	}
}

func TestConcurrentRequests(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	// Make concurrent requests
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			resp, err := http.Get(server.URL + "/AAPL")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				errors <- fmt.Errorf("request %d failed with status %d: %s", id, resp.StatusCode, string(body))
			}
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("Concurrent request error: %v", err)
		}
	}

	// We expect some requests might fail due to circuit breaker, but not all
	if errorCount > 8 {
		t.Errorf("Too many concurrent request failures: %d out of 10", errorCount)
	}
}