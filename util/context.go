package util

import (
	"context"
)

// contextKey is a type for context keys to avoid collisions
// contextKey adalah tipe untuk kunci konteks untuk menghindari tabrakan
type contextKey string

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey contextKey = "trace_id"
	// SpanIDKey is the context key for span ID
	SpanIDKey contextKey = "span_id"
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// SessionIDKey is the context key for session ID
	SessionIDKey contextKey = "session_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// ClientIPKey is the context key for client IP
	ClientIPKey contextKey = "client_ip"
)

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpanID adds span ID to context
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ExtractFromContext extracts all context values - Optimized version
// ExtractFromContext mengekstrak semua nilai konteks - Versi optimal
func ExtractFromContext(ctx context.Context) map[string]string {
	result := GetMapStringFromPool()
	// The defer putMapStringToPool(result) cannot be used here because the map is returned.
	// The caller is responsible for returning the map to the pool.

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		result["trace_id"] = traceID
	}
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
		result["span_id"] = spanID
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		result["user_id"] = userID
	}
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		result["session_id"] = sessionID
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		result["request_id"] = requestID
	}

	return result
}
