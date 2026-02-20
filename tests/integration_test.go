// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/config"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/handlers"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/stock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Setup test environment
	code := m.Run()
	os.Exit(code)
}

func setupTestServer() (*httptest.Server, *config.Config) {
	cfg := &config.Config{
		Port:   "8080",
		Symbol: "MSFT",
		NDays:  7,
		APIKey: "test-key",
	}

	logger := zap.NewNop()
	stockClient := stock.NewClient(cfg, logger)
	handler := handlers.NewHandler(cfg, stockClient, logger)

	r := mux.NewRouter()
	
	// Health endpoints
	r.HandleFunc("/health", handler.HealthHandler).Methods("GET")
	r.HandleFunc("/ready", handler.ReadyHandler).Methods("GET")
	
	// Metrics endpoint
	r.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	
	// Stock endpoints
	r.HandleFunc("/{symbol}", handler.GetStockHandler).Methods("GET")
	r.HandleFunc("/{symbol}/{days}", handler.GetStockHistoryHandler).Methods("GET")
	
	// API docs
	r.HandleFunc("/docs", handler.SwaggerUIHandler).Methods("GET")
	r.HandleFunc("/swagger.yaml", handler.SwaggerHandler).Methods("GET")
	
	server := httptest.NewServer(r)
	return server, cfg
}

func TestHealthEndpoint(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", result["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check for expected metrics
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	metrics := string(body)
	expectedMetrics := []string{
		"go_info",
		"go_goroutines",
		"ping_service_requests_total",
		"ping_service_request_duration_seconds",
		"ping_service_stock_api_duration_seconds",
		"ping_service_cache_hits_total",
		"ping_service_cache_misses_total",
	}

	for _, metric := range expectedMetrics {
		if !contains(metrics, metric) {
			t.Errorf("Expected metric %s not found in response", metric)
		}
	}
}

func TestSwaggerEndpoints(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Test /docs endpoint
	resp, err := http.Get(server.URL + "/docs")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for /docs, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test /swagger.yaml endpoint
	resp, err = http.Get(server.URL + "/swagger.yaml")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for /swagger.yaml, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Make multiple requests to trigger potential circuit breaker behavior
	for i := 0; i < 10; i++ {
		resp, err := http.Get(server.URL + "/TEST")
		if err != nil {
			t.Logf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()
		
		// Should either succeed or fail gracefully
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Unexpected status code %d on request %d", resp.StatusCode, i+1)
		}
	}
}

func TestCachingBehavior(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Make the same request twice to test caching
	symbol := "TEST"
	url := fmt.Sprintf("%s/%s", server.URL, symbol)

	// First request
	resp1, err := http.Get(url)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	resp1.Body.Close()

	// Second request (should be cached)
	resp2, err := http.Get(url)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	resp2.Body.Close()

	// Both should succeed
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("First request failed with status %d", resp1.StatusCode)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Second request failed with status %d", resp2.StatusCode)
	}
}

func TestErrorHandling(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Invalid symbol",
			url:            server.URL + "/INVALID_SYMBOL_12345",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid days parameter",
			url:            server.URL + "/MSFT/invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty symbol",
			url:            server.URL + "/",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(tc.url)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	// Make rapid requests to test rate limiting behavior
	client := &http.Client{Timeout: 5 * time.Second}
	
	for i := 0; i < 20; i++ {
		resp, err := client.Get(server.URL + "/MSFT")
		if err != nil {
			t.Logf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()
		
		// Should not return 429 (Too Many Requests) for reasonable load
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Errorf("Unexpected rate limiting on request %d", i+1)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}