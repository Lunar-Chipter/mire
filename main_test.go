package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"mire/core"
	"mire/formatter"
	"mire/logger"
)

// TestLogEntry is a minimal struct to unmarshal JSON log entries for testing.
type TestLogEntry struct {
	LevelName string                 `json:"level_name"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// TestMainFunction performs an integration test of the main application.
// It captures stdout/stderr and verifies the log file outputs.
func TestMainFunction(t *testing.T) {
	// Clean up log files from previous runs
	os.Remove("app.log")
	os.Remove("errors.log")

	// Backup original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Ensure log files are cleaned up and original stdout/stderr are restored
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		wOut.Close()
		wErr.Close()
		
		// Drain pipes to avoid deadlocks
		var wgDrain sync.WaitGroup
		wgDrain.Add(2)
		go func() { defer wgDrain.Done(); io.Copy(io.Discard, rOut) }()
		go func() { defer wgDrain.Done(); io.Copy(io.Discard, rErr) }()
		wgDrain.Wait()

		rOut.Close()
		rErr.Close()

		// Clean up log files created by main.go
		os.Remove("app.log")
		os.Remove("errors.log")
	}()

	// Run the main function in a separate goroutine to allow capturing output
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
	wg.Wait() // Wait for main to finish

	// Close the writers to flush output to the readers
	wOut.Close()
	wErr.Close()

	// Read captured stdout
	var bufOut bytes.Buffer
	_, err := io.Copy(&bufOut, rOut)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}
	stdout := bufOut.String()
	t.Logf("Captured Stdout:\n%s", stdout)


	// Read captured stderr
	var bufErr bytes.Buffer
	_, err = io.Copy(&bufErr, rErr)
	if err != nil {
		t.Fatalf("Failed to read stderr: %v", err)
	}
	stderr := bufErr.String()
	if stderr != "" {
		t.Errorf("Unexpected content in stderr: %s", stderr)
	}
	t.Logf("Captured Stderr:\n%s", stderr)


	// --- Assertions for Console Output ---
	// Check for presence of key messages in stdout (less strict due to formatting/truncation)
	if !strings.Contains(stdout, "DEMONSTRASI PENGGUNAAN LIBRARY LOGGING MIRE") {
		t.Error("Stdout missing expected header")
	}
	if !strings.Contains(stdout, "Periksa file 'app.log' dan 'errors.log' untuk melihat output log.") {
		t.Error("Stdout missing expected footer log file info")
	}

	// Read and parse app.log with retry
	appLogEntries, err := readLogFileWithRetry(t, "app.log", 2*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to read app.log with retry: %v", err)
	}
	assertLogEntry(t, appLogEntries, "DEBUG", "Pesan debug untuk JSON file logger.")
	assertLogEntryWithField(t, appLogEntries, "INFO", "Transaksi berhasil diproses.", "trans_id", "TXN-001")
	assertLogEntry(t, appLogEntries, "ERROR", "Gagal menyimpan data pengguna ke cache.")

}

// parseJSONLogs parses a byte slice containing multiple JSON objects (one per line)
// into a slice of TestLogEntry.
func parseJSONLogs(t *testing.T, data []byte, entries *[]TestLogEntry) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		var entry TestLogEntry
		if err := decoder.Decode(&entry); err != nil {
			return &wrappedError{
				msg:   "failed to decode JSON object",
				cause: err,
			}
		}

		// Base64 decode the message
		decodedMessage, err := base64.StdEncoding.DecodeString(entry.Message)
		if err != nil {
			return &wrappedError{
				msg:   "failed to base64 decode message for level " + entry.LevelName + ", raw message: " + entry.Message,
				cause: err,
			}
		}
		entry.Message = string(decodedMessage)
		*entries = append(*entries, entry)
	}
	return nil
}

// assertLogEntry checks if a log entry with the given level and message exists.
func assertLogEntry(t *testing.T, entries []TestLogEntry, expectedLevel, expectedMessage string) {
	t.Helper()
	found := false
	for _, entry := range entries {
		if entry.LevelName == expectedLevel && strings.Contains(entry.Message, expectedMessage) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected log entry with level '%s' and message containing '%s' not found.", expectedLevel, expectedMessage)
	}
}

// assertLogEntryWithField checks if a log entry with the given level, message, and a specific field exists.
func assertLogEntryWithField(t *testing.T, entries []TestLogEntry, expectedLevel, expectedMessage, fieldKey string, fieldValue interface{}) {
	t.Helper()
	found := false
	for _, entry := range entries {
		if entry.LevelName == expectedLevel && strings.Contains(entry.Message, expectedMessage) {
			if val, ok := entry.Fields[fieldKey]; ok && val == fieldValue {
				found = true
				break
						}
		}
	}
	if !found {
		t.Errorf("Expected log entry with level '%s', message containing '%s', and field '%s'='%v' not found.", expectedLevel, expectedMessage, fieldKey, fieldValue)
	}
}

// readLogFileWithRetry attempts to read a log file and parse its content as JSON.
// It retries multiple times to account for asynchronous file writing.
func readLogFileWithRetry(t *testing.T, filePath string, timeout, retryInterval time.Duration) ([]TestLogEntry, error) {
	deadline := time.Now().Add(timeout)
	var entries []TestLogEntry

	for time.Now().Before(deadline) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			// If file doesn't exist yet, retry
			if os.IsNotExist(err) {
				time.Sleep(retryInterval)
				continue
			}
			return nil, &wrappedError{
				msg:   "failed to read file " + filePath,
				cause: err,
			}
		}

		if len(content) == 0 {
			time.Sleep(retryInterval)
			continue
		}

		entries = nil // Reset entries for each attempt
		if err := parseJSONLogs(t, content, &entries); err != nil {
			// If parsing failed, assume file is not yet complete and retry
			t.Logf("Failed to parse JSON from %s, retrying: %v", filePath, err)
			time.Sleep(retryInterval)
			continue
		}
		// Successfully parsed, return
		return entries, nil
	}
	return nil, &wrappedError{
		msg:   "timed out reading and parsing JSON from " + filePath,
		cause: nil,
	}
}

// BenchmarkLogInfoDefaultTextFormatter benchmarks the performance of a simple Info log
// using the default TextFormatter to os.Stdout (discarded for benchmark).
func BenchmarkLogInfoDefaultTextFormatter(b *testing.B) {
	// Setup a minimal logger for benchmarking
	cfg := logger.LoggerConfig{
		Level:  core.INFO,
		Output: io.Discard, // Discard output to avoid I/O overhead in benchmark
		Formatter: &formatter.TextFormatter{
			EnableColors:    false, // Disable colors for cleaner benchmark
			ShowTimestamp:   false,
			ShowCaller:      false,
		},
	}
	log := logger.New(cfg)
	defer log.Close() // Ensure logger is closed after benchmark runs

	b.ResetTimer() // Reset timer to exclude setup time

	for i := 0; i < b.N; i++ {
		log.Info("Benchmark message")
	}
}
