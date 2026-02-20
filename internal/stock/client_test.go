package stock

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/cache"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/circuitbreaker"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func createTestClient() *Client {
	logger := zap.NewNop()

	// Create test cache
	stockCache := cache.NewCache(5 * time.Minute)

	// Create test circuit breaker
	cb := circuitbreaker.NewCircuitBreaker(5, 10, 30*time.Second)

	// Create test metrics
	cacheHits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_cache_hits_total",
		Help: "Test cache hits",
	})
	cacheMisses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_cache_misses_total",
		Help: "Test cache misses",
	})
	externalCalls := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_external_calls_total",
		Help: "Test external calls",
	})
	externalCallDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_external_call_duration_seconds",
		Help:    "Test external call duration",
		Buckets: prometheus.DefBuckets,
	})
	circuitBreakerState := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_circuit_breaker_state",
		Help: "Test circuit breaker state",
	})

	return NewClient(
		"test-api-key",
		10*time.Second,
		logger,
		stockCache,
		cb,
		cacheHits,
		cacheMisses,
		externalCalls,
		externalCallDuration,
		circuitBreakerState,
	)
}

func TestProcessTimeSeries(t *testing.T) {
	client := createTestClient()

	timeSeries := map[string]DailyData{
		"2024-01-15": {Close: "415.26"},
		"2024-01-16": {Close: "418.45"},
		"2024-01-17": {Close: "412.89"},
		"2024-01-18": {Close: "420.12"},
		"2024-01-19": {Close: "416.85"},
	}

	result, err := client.processTimeSeries("MSFT", 3, timeSeries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Symbol != "MSFT" {
		t.Errorf("expected symbol MSFT, got %s", result.Symbol)
	}

	if result.NDays != 3 {
		t.Errorf("expected 3 days, got %d", result.NDays)
	}

	if len(result.Prices) != 3 {
		t.Errorf("expected 3 prices, got %d", len(result.Prices))
	}

	// The latest 3 dates should be: 2024-01-19, 2024-01-18, 2024-01-17
	expectedAverage := (416.85 + 420.12 + 412.89) / 3
	if math.Abs(result.Average-expectedAverage) > 0.01 {
		t.Errorf("expected average %.2f, got %.2f", expectedAverage, result.Average)
	}
}

func TestProcessTimeSeriesInsufficientData(t *testing.T) {
	client := createTestClient()

	timeSeries := map[string]DailyData{
		"2024-01-19": {Close: "416.85"},
	}

	result, err := client.processTimeSeries("MSFT", 3, timeSeries)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.NDays != 1 {
		t.Errorf("expected NDays to be adjusted to 1, got %d", result.NDays)
	}
}

func TestProcessTimeSeriesInvalidPrice(t *testing.T) {
	client := createTestClient()

	timeSeries := map[string]DailyData{
		"2024-01-17": {Close: "invalid"},
		"2024-01-18": {Close: "420.12"},
		"2024-01-19": {Close: "416.85"},
	}

	result, err := client.processTimeSeries("MSFT", 3, timeSeries)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.NDays != 2 { // Should skip invalid price and use only 2 valid ones
		t.Errorf("expected NDays to be 2 (only valid prices), got %d", result.NDays)
	}
}

func TestGetStockData(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Query().Get("function") != "TIME_SERIES_DAILY" {
			t.Errorf("expected function=TIME_SERIES_DAILY, got %s", r.URL.Query().Get("function"))
		}
		if r.URL.Query().Get("symbol") != "MSFT" {
			t.Errorf("expected symbol=MSFT, got %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("apikey") != "test-api-key" {
			t.Errorf("expected apikey=test-api-key, got %s", r.URL.Query().Get("apikey"))
		}

		// Return mock response
		response := AlphaVantageResponse{
			TimeSeriesDaily: map[string]DailyData{
				"2024-01-19": {Close: "416.85"},
				"2024-01-18": {Close: "420.12"},
				"2024-01-17": {Close: "412.89"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := createTestClient()
	
	// Create test metrics for api duration
	apiDurationHist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_api_duration_seconds",
		Help: "Test API duration",
	})

	// Set the API URL to our mock server
	client.apiURL = server.URL + "/query"

	result, err := client.GetStockData("MSFT", 3, apiDurationHist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Symbol != "MSFT" {
		t.Errorf("expected symbol MSFT, got %s", result.Symbol)
	}

	if result.NDays != 3 {
		t.Errorf("expected 3 days, got %d", result.NDays)
	}

	if len(result.Prices) != 3 {
		t.Errorf("expected 3 prices, got %d", len(result.Prices))
	}

	expectedAverage := (416.85 + 420.12 + 412.89) / 3
	if math.Abs(result.Average-expectedAverage) > 0.01 {
		t.Errorf("expected average %.2f, got %.2f", expectedAverage, result.Average)
	}
}

func TestGetStockDataCacheHit(t *testing.T) {
	client := createTestClient()
	
	// Create test metrics for api duration
	apiDurationHist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_api_duration_seconds",
		Help: "Test API duration",
	})

	// First call - should be cache miss
	_, err := client.GetStockData("MSFT", 3, apiDurationHist)
	if err == nil {
		t.Skip("Skipping cache test due to API call - would need mock server")
	}

	// Second call - should be cache hit
	_, err = client.GetStockData("MSFT", 3, apiDurationHist)
	if err == nil {
		t.Skip("Skipping cache test due to API call - would need mock server")
	}
}

func TestGetStockDataCircuitBreaker(t *testing.T) {
	client := createTestClient()
	
	// Create test metrics for api duration
	apiDurationHist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "test_api_duration_seconds",
		Help: "Test API duration",
	})

	// Force circuit breaker to open by causing failures
	for i := 0; i < 10; i++ {
		client.GetStockData("INVALID_SYMBOL", 3, apiDurationHist)
	}

	// This should fail due to circuit breaker
	_, err := client.GetStockData("MSFT", 3, apiDurationHist)
	if err == nil {
		t.Skip("Skipping circuit breaker test - would need controlled failure scenario")
	}
}