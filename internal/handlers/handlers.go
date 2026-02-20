package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/config"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/stock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Handler struct {
	config      *config.Config
	stockClient *stock.Client
	logger      *zap.Logger

	// Metrics
	apiRequests  prometheus.Counter
	apiDuration  prometheus.Histogram
}

func NewHandler(cfg *config.Config, stockClient *stock.Client, logger *zap.Logger, apiRequests prometheus.Counter, apiDuration prometheus.Histogram) *Handler {
	return &Handler{
		config:      cfg,
		stockClient: stockClient,
		logger:      logger,
		apiRequests: apiRequests,
		apiDuration: apiDuration,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Health check endpoint
	router.HandleFunc("/health", h.healthHandler).Methods("GET")
	
	// Readiness check endpoint  
	router.HandleFunc("/ready", h.readyHandler).Methods("GET")
	
	// Main stock endpoint
	router.HandleFunc("/", h.stockHandler).Methods("GET")
	
	// Stock symbol endpoint
	router.HandleFunc("/{symbol}", h.stockSymbolHandler).Methods("GET")
	
	// Stock symbol with days endpoint
	router.HandleFunc("/{symbol}/{days}", h.stockSymbolDaysHandler).Methods("GET")
	
	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())
	
	// Documentation
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/index.html")
	})
	router.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.yaml")
	})
}

// Health check endpoint
func (h *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.apiDuration.Observe(time.Since(start).Seconds())
		h.apiRequests.Inc()
	}()

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "stock-service",
		"timestamp": time.Now().Unix(),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// Readiness check endpoint
func (h *Handler) readyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.apiDuration.Observe(time.Since(start).Seconds())
		h.apiRequests.Inc()
	}()

	// Check if we can get basic stock data (using default symbol)
	_, err := h.stockClient.GetStockData(h.config.Symbol, 1, nil)
	if err != nil {
		h.logger.Warn("readiness check failed", zap.Error(err))
		h.sendJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "not ready",
			"error":   "unable to fetch stock data",
			"details": err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":    "ready",
		"service":   "stock-service",
		"timestamp": time.Now().Unix(),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// Main stock endpoint - uses default symbol from config
func (h *Handler) stockHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.apiDuration.Observe(time.Since(start).Seconds())
		h.apiRequests.Inc()
	}()

	h.logger.Info("fetching stock data",
		zap.String("symbol", h.config.Symbol),
		zap.Int("ndays", h.config.NDays))
	
	stockData, err := h.stockClient.GetStockData(h.config.Symbol, h.config.NDays, nil)
	if err != nil {
		h.logger.Error("failed to get stock data", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch stock data", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, stockData)
}

// Stock symbol endpoint - allows dynamic symbol selection
func (h *Handler) stockSymbolHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.apiDuration.Observe(time.Since(start).Seconds())
		h.apiRequests.Inc()
	}()

	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	h.logger.Info("fetching stock data for symbol",
		zap.String("symbol", symbol),
		zap.Int("ndays", h.config.NDays))
	
	stockData, err := h.stockClient.GetStockData(symbol, h.config.NDays, nil)
	if err != nil {
		h.logger.Error("failed to get stock data", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch stock data", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, stockData)
}

// Stock symbol with days endpoint - allows both dynamic symbol and days
func (h *Handler) stockSymbolDaysHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.apiDuration.Observe(time.Since(start).Seconds())
		h.apiRequests.Inc()
	}()

	vars := mux.Vars(r)
	symbol := vars["symbol"]
	daysStr := vars["days"]
	
	// Parse days
	days := h.config.NDays // default to config value
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}
	
	h.logger.Info("fetching stock data for symbol with days",
		zap.String("symbol", symbol),
		zap.Int("ndays", days))
	
	stockData, err := h.stockClient.GetStockData(symbol, days, nil)
	if err != nil {
		h.logger.Error("failed to get stock data", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch stock data", err.Error())
		return
	}
	
	h.sendJSON(w, http.StatusOK, stockData)
}

func (h *Handler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *Handler) sendError(w http.ResponseWriter, statusCode int, message, details string) {
	errorResponse := map[string]interface{}{
		"error":   message,
		"details": details,
		"symbol":  h.config.Symbol,
		"ndays":   h.config.NDays,
	}
	
	h.sendJSON(w, statusCode, errorResponse)
}