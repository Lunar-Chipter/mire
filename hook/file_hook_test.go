package hook

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
)

// TestSimpleFileHookCreation tests creating a new SimpleFileHook
func TestSimpleFileHookCreation(t *testing.T) {
	tempFile := "test_hook.log"
	hook, err := NewFileHook(tempFile)
	
	if err != nil {
		t.Fatalf("NewFileHook failed: %v", err)
	}
	
	if hook == nil {
		t.Fatal("NewFileHook returned nil")
	}
	
	// Check that the correct formatter is used
	_, ok := hook.formatter.(*formatter.JSONFormatter)
	if !ok {
		t.Error("SimpleFileHook should use JSONFormatter")
	}
	
	// Clean up
	hook.Close()
	os.Remove(tempFile)
}

// TestSimpleFileHookCreationError tests error case when file can't be created
func TestSimpleFileHookCreationError(t *testing.T) {
	// Try to create a hook with an invalid path
	hook, err := NewFileHook("/invalid/path/that/does/not/exist/file.log")
	
	if err == nil {
		t.Error("NewFileHook should have failed with invalid path")
		// Close hook if it was created despite the error
		if hook != nil {
			hook.Close()
		}
		return
	}
	
	if hook != nil {
		t.Error("NewFileHook should have returned nil on error")
		hook.Close()
	}
}

// TestSimpleFileHookFire tests firing the hook with different log levels
func TestSimpleFileHookFire(t *testing.T) {
	tempFile := "test_fire_hook.log"
	hook, err := NewFileHook(tempFile)
	if err != nil {
		t.Fatalf("NewFileHook failed: %v", err)
	}
	defer func() {
		hook.Close()
		os.Remove(tempFile)
	}()
	
	// Create a log entry with ERROR level (should be logged)
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.ERROR
	entry.Message = []byte("test error message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp
	
	// Fire the hook - this should write to the file
	err = hook.Fire(entry)
	if err != nil {
		t.Errorf("Fire returned error: %v", err)
	}
	
	// Create a log entry with INFO level (should NOT be logged)
	infoEntry := core.GetEntryFromPool()
	defer core.PutEntryToPool(infoEntry)
	infoEntry.Level = core.INFO
	infoEntry.Message = []byte("test info message")
	infoEntry.Timestamp = core.GetEntryFromPool().Timestamp
	
	// Fire the hook - this should NOT write to the file
	err = hook.Fire(infoEntry)
	if err != nil {
		t.Errorf("Fire returned error for INFO level: %v", err)
	}
}

// TestSimpleFileHookFireError tests error handling when formatter fails
func TestSimpleFileHookFireError(t *testing.T) {
	// Create a mock formatter that always fails
	failingFormatter := &failingTestFormatter{}
	
	tempFile := "test_failing_hook.log"
	// Create hook with failing formatter
	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	hook := &SimpleFileHook{
		writer:    file,
		formatter: failingFormatter,
		file:      file,
	}
	
	// Create a log entry
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.ERROR
	entry.Message = []byte("test message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp
	
	// Fire the hook - this should return an error
	err = hook.Fire(entry)
	if err == nil {
		t.Error("Fire should have returned error with failing formatter")
	} else {
		// Check if the error is wrapped properly
		var wrappedErr *wrappedError
		if errors.As(err, &wrappedErr) {
			if wrappedErr.msg != "file hook failed to format log entry" {
				t.Errorf("Wrapped error message should be 'file hook failed to format log entry', got '%s'", wrappedErr.msg)
			}
		} else {
			t.Error("Fire should have returned wrappedError")
		}
	}
	
	hook.Close()
	os.Remove(tempFile)
}

// TestSimpleFileHookFireWriteError tests error handling when writer fails
func TestSimpleFileHookFireWriteError(t *testing.T) {
	// Create a mock writer that always fails
	failingWriter := &failingTestWriter{}

	hook := &SimpleFileHook{
		writer:    failingWriter,
		formatter: &formatter.JSONFormatter{},
	}

	// Create a log entry with ERROR level
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.ERROR
	entry.Message = []byte("test message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp

	// Fire the hook - this should return an error
	err := hook.Fire(entry)
	if err == nil {
		t.Error("Fire should have returned error with failing writer")
	} else {
		// Check if the error is wrapped properly
		var wrappedErr *wrappedError
		if errors.As(err, &wrappedErr) {
			if wrappedErr.msg != "file hook failed to write log entry" {
				t.Errorf("Wrapped error message should be 'file hook failed to write log entry', got '%s'", wrappedErr.msg)
			}
		} else {
			t.Error("Fire should have returned wrappedError")
		}
	}
}

// TestSimpleFileHookClose tests closing the hook
func TestSimpleFileHookClose(t *testing.T) {
	tempFile := "test_close_hook.log"
	hook, err := NewFileHook(tempFile)
	if err != nil {
		t.Fatalf("NewFileHook failed: %v", err)
	}
	
	// Close the hook
	err = hook.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
	
	// Try to close again - this should not cause issues
	hook.Close()
	// Note: Whether closing an already closed file returns an error depends on the OS/file system
	// For this test, we just ensure it doesn't panic
	
	os.Remove(tempFile)
}

// TestSimpleFileHookCloseNilFile tests closing a hook with nil file
func TestSimpleFileHookCloseNilFile(t *testing.T) {
	hook := &SimpleFileHook{
		file: nil,
	}
	
	// This should not panic
	err := hook.Close()
	if err != nil {
		t.Errorf("Close with nil file returned error: %v", err)
	}
}

// TestWrappedError tests the wrappedError type
func TestWrappedError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := &wrappedError{
		msg:   "wrapped message",
		cause: originalErr,
	}
	
	// Test Error() method
	errorStr := wrapped.Error()
	expected := "wrapped message: original error"
	if errorStr != expected {
		t.Errorf("Error() returned %s, want %s", errorStr, expected)
	}
	
	// Test Unwrap() method
	unwrapped := wrapped.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() returned %v, want %v", unwrapped, originalErr)
	}
	
	// Test with nil cause
	wrappedNil := &wrappedError{
		msg:   "wrapped message",
		cause: nil,
	}
	errorStrNil := wrappedNil.Error()
	expectedNil := "wrapped message"
	if errorStrNil != expectedNil {
		t.Errorf("Error() with nil cause returned %s, want %s", errorStrNil, expectedNil)
	}
}

// failingTestFormatter is a mock formatter that always returns an error
type failingTestFormatter struct{}

func (f *failingTestFormatter) Format(buf *bytes.Buffer, entry *core.LogEntry) error {
	return errors.New("format failed")
}

// failingTestWriter is a mock writer that always returns an error
type failingTestWriter struct{}

func (w *failingTestWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

// TestHookInterfaceImplementation tests that SimpleFileHook implements the Hook interface
func TestHookInterfaceImplementation(t *testing.T) {
	var h Hook
	h = &SimpleFileHook{}
	
	// Create a log entry for testing
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.ERROR
	entry.Message = []byte("test")
	entry.Timestamp = core.GetEntryFromPool().Timestamp
	
	// Create a temporary file for the test
	tempFile := "test_interface_hook.log"
	hook, err := NewFileHook(tempFile)
	if err != nil {
		t.Fatalf("NewFileHook failed: %v", err)
	}
	defer func() {
		hook.Close()
		os.Remove(tempFile)
	}()
	
	// Verify that the Fire method works through the interface
	err = h.Fire(entry)
	if err != nil {
		t.Errorf("Hook.Fire returned error: %v", err)
	}
	
	// Verify that the Close method works through the interface
	err = h.Close()
	if err != nil {
		t.Errorf("Hook.Close returned error: %v", err)
	}
}

// TestSimpleFileHookWithDifferentFormatters tests hook with different formatters
func TestSimpleFileHookWithDifferentFormatters(t *testing.T) {
	tempFile := "test_formatter_hook.log"
	
	// Test with JSON formatter (default)
	hook, err := NewFileHook(tempFile)
	if err != nil {
		t.Fatalf("NewFileHook failed: %v", err)
	}
	defer func() {
		hook.Close()
		os.Remove(tempFile)
	}()
	
	// Verify the formatter is JSON
	_, ok := hook.formatter.(*formatter.JSONFormatter)
	if !ok {
		t.Error("Default formatter should be JSONFormatter")
	}
}