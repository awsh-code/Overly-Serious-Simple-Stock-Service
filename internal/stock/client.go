package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/cache"
	"github.com/awsh-code/Overly-Serious-Simple-Stock-Service/internal/circuitbreaker"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Client struct {
	httpClient          *http.Client
	apiKey              string
	logger              *zap.Logger
	circuitBreaker      *circuitbreaker.CircuitBreaker
	cache               *cache.Cache
	cacheHits           prometheus.Counter
	cacheMisses         prometheus.Counter
	externalCalls       prometheus.Counter
	externalCallDuration prometheus.Histogram
	circuitBreakerState prometheus.Gauge
}

type StockData struct {
	Symbol  string       `json:"symbol"`
	NDays   int          `json:"ndays"`
	Prices  []PricePoint `json:"prices"`
	Average float64      `json:"average"`
}

type PricePoint struct {
	Date  string  `json:"date"`
	Close float64 `json:"close"`
}

type AlphaVantageResponse struct {
	TimeSeriesDaily map[string]DailyData `json:"Time Series (Daily)"`
	Note            string               `json:"Note"`
	ErrorMessage    string               `json:"Error Message"`
}

type DailyData struct {
	Close string `json:"4. close"`
}

func NewClient(
	apiKey string,
	timeout time.Duration,
	logger *zap.Logger,
	cache *cache.Cache,
	circuitBreaker *circuitbreaker.CircuitBreaker,
	cacheHits prometheus.Counter,
	cacheMisses prometheus.Counter,
	externalCalls prometheus.Counter,
	externalCallDuration prometheus.Histogram,
	circuitBreakerState prometheus.Gauge,
) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		apiKey:              apiKey,
		logger:              logger,
		circuitBreaker:      circuitBreaker,
		cache:               cache,
		cacheHits:           cacheHits,
		cacheMisses:         cacheMisses,
		externalCalls:       externalCalls,
		externalCallDuration: externalCallDuration,
		circuitBreakerState: circuitBreakerState,
	}
}

func (c *Client) GetStockData(symbol string, ndays int, apiDurationHist prometheus.Histogram) (*StockData, error) {
	c.logger.Info("fetching stock data", zap.String("symbol", symbol), zap.Int("ndays", ndays))

	// Create cache key
	cacheKey := fmt.Sprintf("%s_%d", symbol, ndays)
	
	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		c.logger.Info("cache hit", zap.String("symbol", symbol), zap.Int("ndays", ndays))
		c.cacheHits.Inc()
		if stockData, ok := cached.(*StockData); ok {
			return stockData, nil
		}
	}

	c.logger.Info("cache miss", zap.String("symbol", symbol), zap.Int("ndays", ndays))
	c.cacheMisses.Inc()

	var result *StockData
	var err error

	cbErr := c.circuitBreaker.Call(func() error {
		result, err = c.fetchStockData(symbol, ndays, apiDurationHist)
		return err
	})

	if cbErr != nil {
		c.logger.Error("circuit breaker error", zap.Error(cbErr))
		return nil, cbErr
	}

	// Cache the successful result
	if err == nil && result != nil {
		c.cache.Set(cacheKey, result)
		c.logger.Info("cached stock data", zap.String("symbol", symbol), zap.Int("ndays", ndays))
	}

	return result, err
}

func (c *Client) fetchStockData(symbol string, ndays int, apiDurationHist prometheus.Histogram) (*StockData, error) {
	start := time.Now()
	defer func() {
		c.externalCallDuration.Observe(time.Since(start).Seconds())
		c.externalCalls.Inc()
	}()

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", symbol, c.apiKey)
	
	c.logger.Info("calling Alpha Vantage API", zap.String("url", url))
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.logger.Error("failed to call Alpha Vantage API", zap.Error(err))
		return nil, fmt.Errorf("failed to call Alpha Vantage API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Alpha Vantage API returned non-200 status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("Alpha Vantage API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var alphaVantageResp AlphaVantageResponse
	if err := json.Unmarshal(body, &alphaVantageResp); err != nil {
		c.logger.Error("failed to unmarshal response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if alphaVantageResp.ErrorMessage != "" {
		c.logger.Error("Alpha Vantage API error", zap.String("error", alphaVantageResp.ErrorMessage))
		return nil, fmt.Errorf("Alpha Vantage API error: %s", alphaVantageResp.ErrorMessage)
	}

	if alphaVantageResp.Note != "" {
		c.logger.Warn("Alpha Vantage API note", zap.String("note", alphaVantageResp.Note))
	}

	if len(alphaVantageResp.TimeSeriesDaily) == 0 {
		c.logger.Error("no time series data returned")
		return nil, fmt.Errorf("no time series data returned")
	}

	return c.processTimeSeries(symbol, ndays, alphaVantageResp.TimeSeriesDaily)
}

func (c *Client) processTimeSeries(symbol string, ndays int, timeSeries map[string]DailyData) (*StockData, error) {
	var dates []string
	for date := range timeSeries {
		dates = append(dates, date)
	}
	
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	
	if len(dates) < ndays {
		ndays = len(dates)
	}
	
	var prices []PricePoint
	var sum float64
	
	for i := 0; i < ndays; i++ {
		date := dates[i]
		dailyData := timeSeries[date]
		
		close, err := parseFloat(dailyData.Close)
		if err != nil {
			c.logger.Error("failed to parse close price", zap.String("date", date), zap.Error(err))
			continue
		}
		
		prices = append(prices, PricePoint{
			Date:  date,
			Close: close,
		})
		sum += close
	}
	
	if len(prices) == 0 {
		return nil, fmt.Errorf("no valid price data found")
	}
	
	average := sum / float64(len(prices))
	
	return &StockData{
		Symbol:  symbol,
		NDays:   len(prices),
		Prices:  prices,
		Average: average,
	}, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}