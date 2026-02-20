package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreakerClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, 5, 30*time.Second)

	// In closed state, all calls should go through
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Error("Expected circuit breaker to allow calls when closed")
	}

	if cb.GetState() != StateClosed {
		t.Error("Expected circuit breaker to remain closed after success")
	}
}

func TestCircuitBreakerOpenState(t *testing.T) {
	cb := NewCircuitBreaker(2, 5, 30*time.Second)

	// Cause failures to open the circuit
	cb.Call(func() error { return errors.New("test error") })
	cb.Call(func() error { return errors.New("test error") })

	if cb.GetState() != StateOpen {
		t.Error("Expected circuit breaker to be open after threshold failures")
	}

	// Should block calls when open
	err := cb.Call(func() error { return nil })
	if err != ErrCircuitBreakerOpen {
		t.Error("Expected circuit breaker to block calls when open")
	}
}

func TestCircuitBreakerHalfOpenState(t *testing.T) {
	cb := NewCircuitBreaker(2, 1, 100*time.Millisecond) // Use 1 for success threshold to close quickly

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return errors.New("test error") })
	}

	if cb.GetState() != StateOpen {
		t.Error("Expected circuit breaker to be open")
	}

	// Wait for timeout to transition to Half-Open
	time.Sleep(150 * time.Millisecond)

	// Should allow one call in half-open state and close immediately (since successThreshold=1)
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Expected circuit breaker to allow calls in half-open state, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected circuit breaker to close after successful half-open call, got %s", cb.GetState())
	}
}

func TestCircuitBreakerHalfOpenToOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 5, 100*time.Millisecond)

	// Open the circuit
	cb.Call(func() error { return errors.New("test error") })

	if cb.GetState() != StateOpen {
		t.Error("Expected circuit breaker to be open")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should fail and go back to open
	err := cb.Call(func() error { return errors.New("test error") })
	if err == nil {
		t.Error("Expected circuit breaker to fail in half-open state")
	}

	if cb.GetState() != StateOpen {
		t.Error("Expected circuit breaker to open after failed half-open call")
	}
}

func TestCircuitBreakerConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(5, 10, 30*time.Second)
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent successes - these should all succeed since circuit starts closed
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cb.Call(func() error { return nil })
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	// Concurrent failures - some of these might cause the circuit to open
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Call(func() error { return errors.New("test error") })
		}()
	}

	wg.Wait()

	// We should have some successful calls
	if successCount == 0 {
		t.Error("Expected at least some successful calls during concurrent access")
	}

	// Verify the circuit breaker is in a valid state
	state := cb.GetState()
	if state != StateClosed && state != StateOpen {
		t.Errorf("Expected circuit breaker to be in a valid state, got %s", state)
	}
}