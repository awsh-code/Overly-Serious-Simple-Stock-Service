package cache

import (
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	key := "test-key"
	value := "test-value"
	cache.Set(key, value)
	retrievedValue, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find key %s", key)
	}
	if retrievedValue != value {
		t.Errorf("Expected value %s, got %s", value, retrievedValue)
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)
	key := "test-key"
	value := "test-value"
	cache.Set(key, value)
	time.Sleep(150 * time.Millisecond) // Wait for expiration
	_, found := cache.Get(key)
	if found {
		t.Errorf("Expected key %s to be expired", key)
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	done := make(chan bool)
	go func() { // Writer goroutine
		for i := 0; i < 100; i++ {
			cache.Set(string(rune(i)), i)
		}
		done <- true
	}()
	go func() { // Reader goroutine
		for i := 0; i < 100; i++ {
			_, _ = cache.Get(string(rune(i)))
		}
		done <- true
	}()
	<-done
	<-done
	t.Log("Concurrent access test passed")
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache(1 * time.Hour)
	key := "test-key"
	value := "test-value"
	cache.Set(key, value)
	
	// Verify it exists
	_, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find key %s before deletion", key)
	}
	
	// Delete it
	cache.Delete(key)
	
	// Verify it's gone
	_, found = cache.Get(key)
	if found {
		t.Errorf("Expected key %s to be deleted", key)
	}
}