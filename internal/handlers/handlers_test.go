package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/awsh-paas-ha/ping-service/internal/config"
	"github.com/awsh-paas-ha/ping-service/internal/stock"
	"github.com/gorilla/mux"
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
		stockClient := &stock.Client{} // Mock this in real tests
		testHandler = NewHandler(cfg, stockClient, logger)
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
	router.HandleFunc("/health", handler.middleware(handler.healthHandler))
	
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	
	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}
	
	if response["service"] != "ping-service" {
		t.Errorf("expected service 'ping-service', got %v", response["service"])
	}
}

func TestMiddleware(t *testing.T) {
	handler, _ := setupTestHandler()
	
	// Test middleware with a simple handler
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	
	middleware := handler.middleware(testHandler)
	middleware(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}