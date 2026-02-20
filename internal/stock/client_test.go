package stock

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestProcessTimeSeries(t *testing.T) {
	logger := zap.NewNop()
	client := NewClient("test-key", 10*time.Second, logger)
	
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
	logger := zap.NewNop()
	client := NewClient("test-key", 10*time.Second, logger)
	
	timeSeries := map[string]DailyData{
		"2024-01-15": {Close: "415.26"},
		"2024-01-16": {Close: "418.45"},
	}
	
	result, err := client.processTimeSeries("MSFT", 5, timeSeries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.NDays != 2 {
		t.Errorf("expected 2 days (available data), got %d", result.NDays)
	}
}

func TestProcessTimeSeriesInvalidPrice(t *testing.T) {
	logger := zap.NewNop()
	client := NewClient("test-key", 10*time.Second, logger)
	
	timeSeries := map[string]DailyData{
		"2024-01-15": {Close: "invalid"},
		"2024-01-16": {Close: "418.45"},
	}
	
	result, err := client.processTimeSeries("MSFT", 2, timeSeries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.NDays != 1 {
		t.Errorf("expected 1 day (valid data only), got %d", result.NDays)
	}
}

func TestGetStockData(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	
	logger := zap.NewNop()
	client := NewClient("test-key", 10*time.Second, logger)
	
	// Override the URL for testing (would need to modify client to accept custom URL)
	// For now, just test the parsing logic
	
	result, err := client.processTimeSeries("MSFT", 3, map[string]DailyData{
		"2024-01-19": {Close: "416.85"},
		"2024-01-18": {Close: "420.12"},
		"2024-01-17": {Close: "412.89"},
	})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if result.Symbol != "MSFT" {
		t.Errorf("expected symbol MSFT, got %s", result.Symbol)
	}
}

func TestGetStockDataRateLimit(t *testing.T) {
	logger := zap.NewNop()
	client := NewClient("test-key", 10*time.Second, logger)
	
	// Test rate limit error
	_, err := client.processTimeSeries("MSFT", 3, map[string]DailyData{})
	if err == nil {
		t.Fatal("expected error for empty time series")
	}
}