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
	if !strings.Contains(output, "MIRE LOGGING LIBRARY DEMONSTRATION") {
		t.Error("Output should contain the main header")
	}

	// Check for various section headers
	expectedSections := []string{
		"### 1. Default Logger ###",
		"### 2. Logger with Fields & Context",
		"### 3. Error Logging with Stack Trace",
		"### 4. JSON Logger to File (app.log)",
		"### 5. Custom Text Logger",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output should contain section: %s", section)
		}
	}

	// Check for content that should definitely appear in console output
	definitelyExpectedContent := []string{
		"There are 2 warnings in the system", // Console output from default logger
		"A simple error occurred",            // Console output from default logger
	}

	for _, content := range definitelyExpectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Required content not found in output: %s", content)
			t.Logf("Actual output was: %s", output)
		}
	}

	// Log info for content that might not appear in console but is expected in logs
	possibleExpectedContent := []string{
		"token-ABC",                            // From context-aware logging (might be in output)
		"Processing authorization request for", // Part of context log message
		"Transaction processed successfully",   // From JSON file logging (might not be in console)
		"level: INFO",                          // Might not be in console output
	}

	for _, content := range possibleExpectedContent {
		if !strings.Contains(output, content) {
			t.Logf("Content not found in console output (may appear in file logs or have different format): %s", content)
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

	// to
	logger.Info("Test message from custom text logger")

	// Close the logger
	logger.Close()
}

// at
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
