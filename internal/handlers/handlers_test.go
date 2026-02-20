package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/config"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/stock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	once     sync.Once
	testHandler *Handler
	testConfig  *config.Config
)

func setupTestHandler() (*Handler, *config.Config) {
	once.Do(func() {
		cfg := &config.Config{
			Port:   "8080",
			Symbol: "MSFT",
			NDays:  7,
			APIKey: "test-key",
		}
		
		logger := zap.NewNop()
		
		// Create a mock stock client that doesn't panic
		stockClient := &stock.Client{}
		
		// Create test metrics
		apiRequests := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_api_requests_total",
			Help: "Test API requests",
		})
		apiDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "test_api_duration_seconds",
			Help: "Test API duration",
		})
		
		testHandler = NewHandler(cfg, stockClient, logger, apiRequests, apiDuration)
		testConfig = cfg
	})
	
	return testHandler, testConfig
}

func TestHealthHandler(t *testing.T) {
	handler, _ := setupTestHandler()
	
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	
	// Create a test router
	router := mux.NewRouter()
	router.HandleFunc("/health", handler.healthHandler)
	
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}