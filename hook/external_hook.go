package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

// ExternalHook is a hook that sends log entries to an external service
type ExternalHook struct {
	endpoint    string
	client      *http.Client
	maxRetries  int
	timeout     time.Duration
	mu          sync.RWMutex
}

// NewExternalHook creates a new ExternalHook that sends logs to the specified endpoint
func NewExternalHook(endpoint string, maxRetries int, timeout time.Duration) (*ExternalHook, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &ExternalHook{
		endpoint:   endpoint,
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		timeout:    timeout,
	}, nil
}

// Fire sends the log entry to the external service with retry logic
func (h *ExternalHook) Fire(entry *core.LogEntry) error {
	// Serialize the log entry to JSON
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Create HTTP request context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	// Try to send with exponential backoff
	var lastErr error
	for attempt := 0; attempt < h.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", h.endpoint, bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to send request: %w", attempt+1, err)
			
			// Exponential backoff: wait 2^attempt * 100ms
			if attempt < h.maxRetries-1 {
				backoffDuration := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
				time.Sleep(backoffDuration)
			}
			continue
		}

		// Read response body to prevent connection leaks
		_, _ = io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // Success
		}

		lastErr = fmt.Errorf("attempt %d: service returned status %d", attempt+1, resp.StatusCode)
		
		// Exponential backoff
		if attempt < h.maxRetries-1 {
			backoffDuration := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			time.Sleep(backoffDuration)
		}
	}

	return fmt.Errorf("failed to send log after %d attempts: %w", h.maxRetries, lastErr)
}

// Close closes the external hook (cleanup resources if any)
func (h *ExternalHook) Close() error {
	// In this implementation, there's no specific cleanup needed
	// The HTTP client doesn't need to be closed
	return nil
}