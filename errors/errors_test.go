package errors

import (
	"bytes"
	"testing"
)

// TestInvalidLogLevelErrorCreation tests the creation and pooling of InvalidLogLevelError
func TestInvalidLogLevelErrorCreation(t *testing.T) {
	err := NewInvalidLogLevelError("invalid_level")
	if err == nil {
		t.Fatal("NewInvalidLogLevelError returned nil")
	}

	if err.level != "invalid_level" {
		t.Errorf("InvalidLogLevelError.level = %s, want invalid_level", err.level)
	}

	// Test that the buffer is empty initially
	if err.buf.Len() != 0 {
		t.Error("InvalidLogLevelError.buf should be empty initially")
	}

	// Return to pool
	PutInvalidLogLevelError(err)

	// Get another error from pool to test reuse
	err2 := NewInvalidLogLevelError("another_invalid")
	if err2 == nil {
		t.Fatal("NewInvalidLogLevelError after pool return returned nil")
	}

	PutInvalidLogLevelError(err2)
}

// TestInvalidLogLevelErrorAppendError tests the AppendError method
func TestInvalidLogLevelErrorAppendError(t *testing.T) {
	err := NewInvalidLogLevelError("bad_level")
	defer PutInvalidLogLevelError(err)

	buf := new(bytes.Buffer)
	err.AppendError(buf)

	result := buf.String()
	expected := "invalid log level: bad_level"
	if result != expected {
		t.Errorf("AppendError wrote %s, want %s", result, expected)
	}
}

// TestInvalidLogLevelErrorError tests the Error method
func TestInvalidLogLevelErrorError(t *testing.T) {
	err := NewInvalidLogLevelError("bad_level")
	defer PutInvalidLogLevelError(err)

	errorStr := err.Error()
	expected := "invalid log level: bad_level"
	if errorStr != expected {
		t.Errorf("Error() returned %s, want %s", errorStr, expected)
	}
}

// TestCustomError tests the customError type
func TestCustomError(t *testing.T) {
	err := &customError{msg: "test error"}
	
	if err.Error() != "test error" {
		t.Errorf("customError.Error() returned %s, want test error", err.Error())
	}
}

// TestErrAsyncBufferFull tests the ErrAsyncBufferFull variable
func TestErrAsyncBufferFull(t *testing.T) {
	if ErrAsyncBufferFull == nil {
		t.Error("ErrAsyncBufferFull is nil")
	}
	
	if ErrAsyncBufferFull.Error() != "async log channel full" {
		t.Errorf("ErrAsyncBufferFull.Error() returned %s, want 'async log channel full'", ErrAsyncBufferFull.Error())
	}
}

// TestInvalidLogLevelErrorConcurrent tests the InvalidLogLevelError in a concurrent context
func TestInvalidLogLevelErrorConcurrent(t *testing.T) {
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			err := NewInvalidLogLevelError("concurrent_test")
			if err == nil {
				t.Error("NewInvalidLogLevelError returned nil in concurrent test")
			}
			
			buf := new(bytes.Buffer)
			err.AppendError(buf)
			result := buf.String()
			if result != "invalid log level: concurrent_test" {
				t.Errorf("Concurrent AppendError returned %s, want 'invalid log level: concurrent_test'", result)
			}
			
			PutInvalidLogLevelError(err)
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestInvalidLogLevelErrorPoolReuse tests that the pool properly reuses objects
func TestInvalidLogLevelErrorPoolReuse(t *testing.T) {
	// Get an error from the pool
	err1 := NewInvalidLogLevelError("first")
	if err1.level != "first" {
		t.Errorf("First error should have level 'first', got %s", err1.level)
	}
	
	// Return it to the pool
	PutInvalidLogLevelError(err1)
	
	// Get another error from the pool
	err2 := NewInvalidLogLevelError("second")
	if err2.level != "second" {
		t.Errorf("Second error should have level 'second', got %s", err2.level)
	}
	
	// The internal buffer should have been reset
	if err2.buf.Len() != 0 {
		t.Error("Buffer should have been reset when error was returned to pool")
	}
	
	PutInvalidLogLevelError(err2)
}

// TestInvalidLogLevelErrorImplementsErrorAppender tests that InvalidLogLevelError implements ErrorAppender
func TestInvalidLogLevelErrorImplementsErrorAppender(t *testing.T) {
	err := NewInvalidLogLevelError("test")
	defer PutInvalidLogLevelError(err)
	
	// This should compile without error if the interface is implemented
	var appender interface{} = err
	_, ok := appender.(interface{ AppendError(*bytes.Buffer) })
	if !ok {
		t.Error("InvalidLogLevelError does not implement AppendError method")
	}
}

// TestInvalidLogLevelErrorErrorInterface tests that InvalidLogLevelError implements the standard error interface
func TestInvalidLogLevelErrorErrorInterface(t *testing.T) {
	err := NewInvalidLogLevelError("test")
	defer PutInvalidLogLevelError(err)
	
	// This should compile without error if the error interface is implemented
	var stdErr error = err
	if stdErr.Error() == "" {
		t.Error("InvalidLogLevelError does not properly implement the error interface")
	}
}