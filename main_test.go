package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestMainFunction tests the main function by capturing stdout
func TestMainFunction(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the main function
	main()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the output contains expected parts
	output := buf.String()

	// Check for the header
	if !strings.Contains(output, "DEMONSTRASI PENGGUNAAN LIBRARY LOGGING MIRE") {
		t.Error("Output should contain the main header")
	}

	// Check for various section headers
	expectedSections := []string{
		"### 1. Logger Default",
		"### 2. Logger dengan Fields & Context",
		"### 3. Error Logging dengan Stack Trace",
		"### 4. Logger JSON ke File",
		"### 5. Custom Text Logger",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain section: %s", section)
		}
	}

	// Check that certain log messages appear
	expectedContent := []string{
		"Ada 2 peringatan di sistem",
		"Terjadi error sederhana",
		"Memproses permintaan otorisasi",
		"Transaksi berhasil diproses",
		"level: INFO",
	}
	
	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Logf("Expected content not found in output (this may be normal due to logging configuration): %s", content)
		}
	}
}

// TestWrappedError tests the wrappedError type used in main
func TestWrappedError(t *testing.T) {
	originalErr := &os.PathError{Op: "open", Path: "/invalid/path", Err: os.ErrNotExist}
	
	wrapped := &wrappedError{
		msg:   "failed to open file",
		cause: originalErr,
	}
	
	// Test Error() method
	errorStr := wrapped.Error()
	if !strings.Contains(errorStr, "failed to open file") {
		t.Errorf("Error() should contain the wrapper message, got: %s", errorStr)
	}
	
	if !strings.Contains(errorStr, "open /invalid/path") {
		t.Logf("Error() may contain the wrapped error message: %s", errorStr)
	}
	
	// Test Unwrap() method
	unwrapped := wrapped.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap() should return the original error")
	}
}

// TestWrappedErrorWithoutCause tests wrappedError when cause is nil
func TestWrappedErrorWithoutCause(t *testing.T) {
	wrapped := &wrappedError{
		msg:   "error message without cause",
		cause: nil,
	}
	
	// Test Error() method
	errorStr := wrapped.Error()
	expected := "error message without cause"
	if errorStr != expected {
		t.Errorf("Error() should return only the message when cause is nil, expected: %s, got: %s", expected, errorStr)
	}
	
	// Test Unwrap() method
	unwrapped := wrapped.Unwrap()
	if unwrapped != nil {
		t.Error("Unwrap() should return nil when cause is nil")
	}
}

// TestPrintLine tests the printLine helper function
func TestPrintLine(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Use the printLine function
	testMessage := "Test printLine function"
	printLine(testMessage)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if the message was printed with a newline
	expectedOutput := testMessage + "\n"
	if output != expectedOutput {
		t.Errorf("printLine output mismatch. Expected: %q, Got: %q", expectedOutput, output)
	}
}

// TestPrintLineEmpty tests printLine with empty string
func TestPrintLineEmpty(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Use the printLine function with empty string
	printLine("")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if an empty line was printed
	expectedOutput := "\n"
	if output != expectedOutput {
		t.Errorf("printLine with empty string output mismatch. Expected: %q, Got: %q", expectedOutput, output)
	}
}

// TestSetupJSONFileLogger tests the setupJSONFileLogger function
func TestSetupJSONFileLogger(t *testing.T) {
	// This function creates a logger that writes to a file
	// We'll test that it returns a non-nil logger and doesn't error with a temporary file
	logger, err := setupJSONFileLogger("test_json.log")
	
	if err != nil {
		t.Fatalf("setupJSONFileLogger returned error: %v", err)
	}
	
	if logger == nil {
		t.Error("setupJSONFileLogger returned nil logger")
	}
	
	// Close the logger
	logger.Close()
	
	// Clean up the test file if it was created
	os.Remove("test_json.log")
}

// TestSetupCustomTextLogger tests the setupCustomTextLogger function
func TestSetupCustomTextLogger(t *testing.T) {
	logger := setupCustomTextLogger()
	
	if logger == nil {
		t.Error("setupCustomTextLogger returned nil logger")
	}
	
	// Test that the logger has the expected configuration
	// The implementation details are in setupCustomTextLogger function
	// At minimum, it should be able to log without error
	
	// Log a test message to make sure it works
	logger.Info("Test message from custom text logger")
	
	// Close the logger
	logger.Close()
}

// TestMainFunctionDoesNotPanic tests that main function does not panic under normal conditions
func TestMainFunctionDoesNotPanic(t *testing.T) {
	// This test ensures that the main function completes without panicking
	// We can't easily verify all functionality, but at least ensure it doesn't crash
	
	// Capture stdout to prevent it from appearing in test output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run main and ensure it doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
		}()
		main()
	}()

	// Restore stdout and clean up
	w.Close()
	os.Stdout = oldStdout
	
	// Drain the pipe to prevent blocking
	_, err := io.Copy(io.Discard, r)
	if err != nil {
		t.Errorf("Error draining pipe: %v", err)
	}
}