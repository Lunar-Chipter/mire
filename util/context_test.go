package util

import (
	"context"
	"testing"
)

// TestWithTraceID tests adding trace ID to context
func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "test-trace-id"
	
	newCtx := WithTraceID(ctx, traceID)
	
	// Extract the trace ID from the context
	extractedTraceID, ok := newCtx.Value(TraceIDKey).(string)
	if !ok {
		t.Fatal("Trace ID was not set properly in context")
	}
	
	if extractedTraceID != traceID {
		t.Errorf("Expected trace ID %s, got %s", traceID, extractedTraceID)
	}
	
	// Verify original context is unchanged
	originalTraceID, ok := ctx.Value(TraceIDKey).(string)
	if ok {
		t.Errorf("Original context should not have trace ID, but got %s", originalTraceID)
	}
}

// TestWithSpanID tests adding span ID to context
func TestWithSpanID(t *testing.T) {
	ctx := context.Background()
	spanID := "test-span-id"
	
	newCtx := WithSpanID(ctx, spanID)
	
	// Extract the span ID from the context
	extractedSpanID, ok := newCtx.Value(SpanIDKey).(string)
	if !ok {
		t.Fatal("Span ID was not set properly in context")
	}
	
	if extractedSpanID != spanID {
		t.Errorf("Expected span ID %s, got %s", spanID, extractedSpanID)
	}
	
	// Verify original context is unchanged
	originalSpanID, ok := ctx.Value(SpanIDKey).(string)
	if ok {
		t.Errorf("Original context should not have span ID, but got %s", originalSpanID)
	}
}

// TestWithUserID tests adding user ID to context
func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-id"
	
	newCtx := WithUserID(ctx, userID)
	
	// Extract the user ID from the context
	extractedUserID, ok := newCtx.Value(UserIDKey).(string)
	if !ok {
		t.Fatal("User ID was not set properly in context")
	}
	
	if extractedUserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, extractedUserID)
	}
	
	// Verify original context is unchanged
	originalUserID, ok := ctx.Value(UserIDKey).(string)
	if ok {
		t.Errorf("Original context should not have user ID, but got %s", originalUserID)
	}
}

// TestWithSessionID tests adding session ID to context
func TestWithSessionID(t *testing.T) {
	ctx := context.Background()
	sessionID := "test-session-id"
	
	newCtx := WithSessionID(ctx, sessionID)
	
	// Extract the session ID from the context
	extractedSessionID, ok := newCtx.Value(SessionIDKey).(string)
	if !ok {
		t.Fatal("Session ID was not set properly in context")
	}
	
	if extractedSessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, extractedSessionID)
	}
	
	// Verify original context is unchanged
	originalSessionID, ok := ctx.Value(SessionIDKey).(string)
	if ok {
		t.Errorf("Original context should not have session ID, but got %s", originalSessionID)
	}
}

// TestWithRequestID tests adding request ID to context
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-id"
	
	newCtx := WithRequestID(ctx, requestID)
	
	// Extract the request ID from the context
	extractedRequestID, ok := newCtx.Value(RequestIDKey).(string)
	if !ok {
		t.Fatal("Request ID was not set properly in context")
	}
	
	if extractedRequestID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, extractedRequestID)
	}
	
	// Verify original context is unchanged
	originalRequestID, ok := ctx.Value(RequestIDKey).(string)
	if ok {
		t.Errorf("Original context should not have request ID, but got %s", originalRequestID)
	}
}


// TestExtractFromContext tests extracting all context values
func TestExtractFromContext(t *testing.T) {
	ctx := context.Background()

	// Add all possible context values that exist in the original code
	traceID := "extract-trace-id"
	spanID := "extract-span-id"
	userID := "extract-user-id"
	sessionID := "extract-session-id"
	requestID := "extract-request-id"

	ctx = WithTraceID(ctx, traceID)
	ctx = WithSpanID(ctx, spanID)
	ctx = WithUserID(ctx, userID)
	ctx = WithSessionID(ctx, sessionID)
	ctx = WithRequestID(ctx, requestID)

	// Extract all values using ExtractFromContext
	result := ExtractFromContext(ctx)
	defer PutMapStringToPool(result) // Important: return the map to the pool

	// Verify all values are present
	if result["trace_id"] != traceID {
		t.Errorf("Expected trace_id %s, got %s", traceID, result["trace_id"])
	}

	if result["span_id"] != spanID {
		t.Errorf("Expected span_id %s, got %s", spanID, result["span_id"])
	}

	if result["user_id"] != userID {
		t.Errorf("Expected user_id %s, got %s", userID, result["user_id"])
	}

	if result["session_id"] != sessionID {
		t.Errorf("Expected session_id %s, got %s", sessionID, result["session_id"])
	}

	if result["request_id"] != requestID {
		t.Errorf("Expected request_id %s, got %s", requestID, result["request_id"])
	}
}

// TestExtractFromContextWithMissingValues tests extracting context values when some are missing
func TestExtractFromContextWithMissingValues(t *testing.T) {
	ctx := context.Background()

	// Add only some values
	traceID := "partial-trace-id"
	userID := "partial-user-id"

	ctx = WithTraceID(ctx, traceID)
	ctx = WithUserID(ctx, userID)
	// Skip adding other values

	// Extract values
	result := ExtractFromContext(ctx)
	defer PutMapStringToPool(result) // Important: return the map to the pool

	// Should have trace_id and user_id
	if result["trace_id"] != traceID {
		t.Errorf("Expected trace_id %s, got %s", traceID, result["trace_id"])
	}

	if result["user_id"] != userID {
		t.Errorf("Expected user_id %s, got %s", userID, result["user_id"])
	}

	// Should not have other values
	if _, exists := result["span_id"]; exists {
		t.Error("Result should not contain span_id when not set in context")
	}

	if _, exists := result["session_id"]; exists {
		t.Error("Result should not contain session_id when not set in context")
	}

	if _, exists := result["request_id"]; exists {
		t.Error("Result should not contain request_id when not set in context")
	}
}

// TestExtractFromContextEmpty tests extracting from an empty context
func TestExtractFromContextEmpty(t *testing.T) {
	ctx := context.Background()
	
	// Extract from empty context
	result := ExtractFromContext(ctx)
	defer PutMapStringToPool(result) // Important: return the map to the pool
	
	// Should be an empty map
	if len(result) != 0 {
		t.Errorf("Expected empty map, got map with %d keys", len(result))
	}
}

// TestContextKeyString tests that context keys are properly defined as strings
func TestContextKeyString(t *testing.T) {
	// Check that all keys are string type
	_ = string(TraceIDKey)
	_ = string(SpanIDKey)
	_ = string(UserIDKey)
	_ = string(SessionIDKey)
	_ = string(RequestIDKey)

	// Check that keys are all different
	keys := []contextKey{
		TraceIDKey,
		SpanIDKey,
		UserIDKey,
		SessionIDKey,
		RequestIDKey,
	}

	// Make sure all keys are unique
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] == keys[j] {
				t.Errorf("Context keys %v and %v are the same", keys[i], keys[j])
			}
		}
	}
}

// TestExtractFromContextWithEmptyStringValues tests extracting context values when values are empty strings
func TestExtractFromContextWithEmptyStringValues(t *testing.T) {
	ctx := context.Background()
	
	// Add empty string values
	ctx = WithTraceID(ctx, "")
	ctx = WithUserID(ctx, "")
	ctx = WithRequestID(ctx, "")
	
	// Extract values
	result := ExtractFromContext(ctx)
	defer PutMapStringToPool(result) // Important: return the map to the pool
	
	// Empty string values should not be included in the result
	// according to the implementation logic
	for key, value := range result {
		if key == "trace_id" || key == "user_id" || key == "request_id" {
			if value != "" {
				t.Errorf("Expected empty string for %s, got %s", key, value)
			}
		}
	}
}

// TestContextValueTypes tests that context values are stored and retrieved as strings
func TestContextValueTypes(t *testing.T) {
	ctx := context.Background()
	
	// Add a value using WithTraceID (which stores as string)
	ctx = WithTraceID(ctx, "test-id")
	
	// Extract and verify it's a string
	rawValue := ctx.Value(TraceIDKey)
	if rawValue == nil {
		t.Fatal("Value not found in context")
	}
	
	if _, ok := rawValue.(string); !ok {
		t.Errorf("Expected value to be string, got %T", rawValue)
	}
}