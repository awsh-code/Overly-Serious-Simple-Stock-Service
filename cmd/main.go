package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/cache"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/circuitbreaker"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/config"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/handlers"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/middleware"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/stock"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	// Create cache
	stockCache := cache.New(cfg.CacheTTL)

	// Create circuit breaker
	cb := circuitbreaker.New(cfg.CircuitBreakerTimeout, logger)

	// Create Prometheus metrics
	cacheHits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stock_api_cache_hits_total",
		Help: "Total number of cache hits",
	})
	cacheMisses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stock_api_cache_misses_total",
		Help: "Total number of cache misses",
	})
	externalCalls := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stock_api_external_calls_total",
		Help: "Total number of external API calls",
	})
	externalCallDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "stock_api_external_call_duration_seconds",
		Help:    "Duration of external API calls in seconds",
		Buckets: prometheus.DefBuckets,
	})
	circuitBreakerState := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stock_api_circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
	})
	apiRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stock_api_requests_total",
		Help: "Total number of API requests",
	})
	apiDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "stock_api_request_duration_seconds",
		Help:    "Duration of API requests in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// Register all metrics
	prometheus.MustRegister(
		cacheHits,
		cacheMisses,
		externalCalls,
		externalCallDuration,
		circuitBreakerState,
		apiRequests,
		apiDuration,
	)

	// Create stock client with all dependencies
	stockClient := stock.NewClient(
		cfg.APIKey,
		cfg.APITimeout,
		logger,
		stockCache,
		cb,
		cacheHits,
		cacheMisses,
		externalCalls,
		externalCallDuration,
		circuitBreakerState,
	)

	// Create handler
	handler := handlers.NewHandler(cfg, stockClient, logger, apiRequests, apiDuration)

	logger.Info("Starting Overly-Serious-Simple-Stock-Service",
		zap.String("symbol", cfg.Symbol),
		zap.Int("ndays", cfg.NDays),
		zap.String("port", cfg.Port),
	)

	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Metrics)

	// Register routes
	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
	}

	go func() {
		logger.Info("starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}
	logger.Info("server exited gracefully")
}