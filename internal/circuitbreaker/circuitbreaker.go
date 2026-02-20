package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrCircuitBreakerHalfOpen = errors.New("circuit breaker is half-open")
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	failureThreshold   int
	successThreshold   int
	timeout            time.Duration
	failureCount       int
	successCount       int
	lastFailureTime    time.Time
	state              State
	mu                 sync.Mutex
}

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		state:            StateClosed,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from Open to Half-Open
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = StateHalfOpen
		cb.failureCount = 0
		cb.successCount = 0
	}

	switch cb.state {
	case StateOpen:
		return ErrCircuitBreakerOpen
	case StateHalfOpen:
		err := fn()
		if err != nil {
			cb.failureCount++
			if cb.failureCount >= cb.failureThreshold {
				cb.state = StateOpen
				cb.lastFailureTime = time.Now()
			}
			return err
		} else {
			cb.successCount++
			if cb.successCount >= cb.successThreshold {
				cb.state = StateClosed
				cb.failureCount = 0
				cb.successCount = 0
			}
			return nil
		}
	case StateClosed:
		err := fn()
		if err != nil {
			cb.failureCount++
			cb.successCount = 0
			if cb.failureCount >= cb.failureThreshold {
				cb.state = StateOpen
				cb.lastFailureTime = time.Now()
			}
			return err
		} else {
			cb.failureCount = 0
			cb.successCount = 0
			return nil
		}
	}

	return nil
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) GetMetrics() (state State, failureCount, successCount int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state, cb.failureCount, cb.successCount
}