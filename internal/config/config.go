package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                      string
	Symbol                    string
	NDays                     int
	APIKey                    string
	ServerReadTimeout         time.Duration
	ServerWriteTimeout        time.Duration
	APITimeout                time.Duration
	CacheTTL                  time.Duration
	CircuitBreakerTimeout     time.Duration
	CircuitBreakerThreshold   int
	CircuitBreakerSuccessThreshold int
}

func Load() *Config {
	ndays, _ := strconv.Atoi(getEnv("NDAYS", "7"))
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL", "300"))
	circuitBreakerTimeout, _ := strconv.Atoi(getEnv("CIRCUIT_BREAKER_TIMEOUT", "30"))
	circuitBreakerThreshold, _ := strconv.Atoi(getEnv("CIRCUIT_BREAKER_THRESHOLD", "5"))
	circuitBreakerSuccessThreshold, _ := strconv.Atoi(getEnv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD", "10"))
	
	return &Config{
		Port:                      getEnv("PORT", "8080"),
		Symbol:                    getEnv("SYMBOL", "MSFT"),
		NDays:                     ndays,
		APIKey:                    getEnv("APIKEY", "demo"),
		ServerReadTimeout:         15 * time.Second,
		ServerWriteTimeout:        15 * time.Second,
		APITimeout:                10 * time.Second,
		CacheTTL:                  time.Duration(cacheTTL) * time.Second,
		CircuitBreakerTimeout:     time.Duration(circuitBreakerTimeout) * time.Second,
		CircuitBreakerThreshold:   circuitBreakerThreshold,
		CircuitBreakerSuccessThreshold: circuitBreakerSuccessThreshold,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}