package errors

import (
	"bytes"
	"testing"
)

// TestInvalidLevelCreation tests the creation and pooling of InvalidLevel
func TestInvalidLevelCreation(t *testing.T) {
	err := NewInvalidLevel("invalid_level")
	if err == nil {
		t.Fatal("NewInvalidLevel returned nil")
	}

	if err.level != "invalid_level" {
		t.Errorf("InvalidLevel.level = %s, want invalid_level", err.level)
	}

	// Test that the buffer is empty initially
	if err.buf.Len() != 0 {
		t.Error("InvalidLevel.buf should be empty initially")
	}

	// Return to pool
	PutInvalidLevel(err)

	// Get another error from pool to test reuse
	err2 := NewInvalidLevel("another_invalid")
	if err2 == nil {
		t.Fatal("NewInvalidLevel after pool return returned nil")
	}

	PutInvalidLevel(err2)
}

// TestInvalidLevelAppendError tests the AppendError method
func TestInvalidLevelAppendError(t *testing.T) {
	err := NewInvalidLevel("bad_level")
	defer PutInvalidLevel(err)

	buf := new(bytes.Buffer)
	err.AppendError(buf)

	result := buf.String()
	expected := "invalid log level: bad_level"
	if result != expected {
		t.Errorf("AppendError wrote %s, want %s", result, expected)
	}
}

// TestInvalidLevelError tests the Error method
func TestInvalidLevelError(t *testing.T) {
	err := NewInvalidLevel("bad_level")
	defer PutInvalidLevel(err)

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

// TestInvalidLevelConcurrent tests the InvalidLevel in a concurrent context
func TestInvalidLevelConcurrent(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			err := NewInvalidLevel("concurrent_test")
			if err == nil {
				t.Error("NewInvalidLevel returned nil in concurrent test")
			}

			buf := new(bytes.Buffer)
			err.AppendError(buf)
			result := buf.String()
			if result != "invalid log level: concurrent_test" {
				t.Errorf("Concurrent AppendError returned %s, want 'invalid log level: concurrent_test'", result)
			}

			PutInvalidLevel(err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestInvalidLevelPoolReuse tests that the pool properly reuses objects
func TestInvalidLevelPoolReuse(t *testing.T) {
	// Get an error from the pool
	err1 := NewInvalidLevel("first")
	if err1.level != "first" {
		t.Errorf("First error should have level 'first', got %s", err1.level)
	}

	// Return it to the pool
	PutInvalidLevel(err1)

	// Get another error from the pool
	err2 := NewInvalidLevel("second")
	if err2.level != "second" {
		t.Errorf("Second error should have level 'second', got %s", err2.level)
	}

	// The internal buffer should have been reset
	if err2.buf.Len() != 0 {
		t.Error("Buffer should have been reset when error was returned to pool")
	}

	PutInvalidLevel(err2)
}

// TestInvalidLevelImplementsErrorAppender tests that InvalidLevel implements ErrorAppender
func TestInvalidLevelImplementsErrorAppender(t *testing.T) {
	err := NewInvalidLevel("test")
	defer PutInvalidLevel(err)

	// This should compile without error if the interface is implemented
	var appender interface{} = err
	_, ok := appender.(interface{ AppendError(*bytes.Buffer) })
	if !ok {
		t.Error("InvalidLevel does not implement AppendError method")
	}
}

// TestInvalidLevelErrorInterface tests that InvalidLevel implements the standard error interface
func TestInvalidLevelErrorInterface(t *testing.T) {
	err := NewInvalidLevel("test")
	defer PutInvalidLevel(err)

	// This should compile without error if the error interface is implemented
	var stdErr error = err
	if stdErr.Error() == "" {
		t.Error("InvalidLevel does not properly implement the error interface")
	}
}
