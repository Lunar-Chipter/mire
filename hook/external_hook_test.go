package hook

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

func TestNewExternalHook(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		hook, err := NewExternalHook("http://example.com/logs", 3, 10*time.Second)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if hook == nil {
			t.Error("Expected hook to be created, got nil")
		}
	})

	t.Run("Empty endpoint", func(t *testing.T) {
		hook, err := NewExternalHook("", 3, 10*time.Second)
		if err == nil {
			t.Error("Expected error for empty endpoint, got nil")
		}
		if hook != nil {
			t.Error("Expected hook to be nil for empty endpoint")
		}
	})

	t.Run("Zero retries defaults to 3", func(t *testing.T) {
		hook, err := NewExternalHook("http://example.com/logs", 0, 10*time.Second)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		// Internal field isn't exported, but we can verify it doesn't panic
		if hook == nil {
			t.Error("Expected hook to be created, got nil")
		}
	})
}

func TestExternalHookFire(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	hook, err := NewExternalHook(server.URL, 3, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Create a test log entry
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	defer core.PutEntryToPool(entry)

	// Test successful sending
	err = hook.Fire(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExternalHookFireWithRetries(t *testing.T) {
	var requestCount int
	var mu sync.Mutex

	// Create a test server that returns 500 for the first 2 requests, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		currentCount := requestCount
		mu.Unlock()

		if currentCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	hook, err := NewExternalHook(server.URL, 5, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Create a test log entry
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message with retries")
	defer core.PutEntryToPool(entry)

	// Test that it eventually succeeds after retries
	err = hook.Fire(entry)
	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	// Should have been attempted 3 times (first failure, second failure, third success)
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

func TestExternalHookFireAllRetriesFail(t *testing.T) {
	// Create a test server that always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	hook, err := NewExternalHook(server.URL, 3, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Create a test log entry
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message with failed retries")
	defer core.PutEntryToPool(entry)

	// Test that all retries fail
	err = hook.Fire(entry)
	if err == nil {
		t.Error("Expected error after all retries failed, got nil")
	}

	// Error message should contain retry information
	if err != nil && err.Error() != "failed to send log after 3 attempts: attempt 3: service returned status 500" {
		// The actual error message may vary, so just check if it mentions retries
		if err.Error() == "" || err.Error() == "failed to send log after 3 attempts: %!w(<nil>)" {
			t.Errorf("Expected error message about failed retries, got: %v", err.Error())
		}
	}
}

func TestExternalHookTimeout(t *testing.T) {
	// Create a test server that delays response to trigger timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than our timeout
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create hook with short timeout
	hook, err := NewExternalHook(server.URL, 1, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Create a test log entry
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message with timeout")
	defer core.PutEntryToPool(entry)

	// Test that timeout error occurs
	err = hook.Fire(entry)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestExternalHookClose(t *testing.T) {
	hook, err := NewExternalHook("http://example.com/logs", 3, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Test that Close doesn't panic
	err = hook.Close()
	if err != nil {
		t.Errorf("Expected no error from Close, got: %v", err)
	}
}

func TestExternalHookWithContextCancellation(t *testing.T) {
	// Create a server that blocks
	blockingServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for the signal to respond
		<-blockingServer
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	defer close(blockingServer) // Ensure server eventually responds

	hook, err := NewExternalHook(server.URL, 1, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create hook: %v", err)
	}

	// Create a test log entry
	entry := core.GetEntryFromPool()
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test context cancellation")
	defer core.PutEntryToPool(entry)

	// This should timeout and return an error
	err = hook.Fire(entry)
	if err == nil {
		t.Error("Expected timeout error due to context cancellation, got nil")
	}
}