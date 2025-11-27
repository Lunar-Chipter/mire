package hook

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
)

// Hook interface defines the contract for log processing hooks
type Hook interface {
	Fire(entry *core.LogEntry) error
	Close() error
}

// wrappedError wraps an error with a message
type wrappedError struct {
	msg   string
	cause error
}

func (e *wrappedError) Error() string {
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.cause
}

// SimpleFileHook is a hook that writes log entries to a file.
type SimpleFileHook struct {
	mu        sync.Mutex
	writer    io.Writer
	formatter formatter.Formatter
	file      *os.File // Keep reference to the file to close it
}

// NewFileHook creates a new SimpleFileHook that writes to the specified file.
func NewFileHook(filePath string) (*SimpleFileHook, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, &wrappedError{
			msg:   "failed to open log file " + filePath + " for hook",
			cause: err,
		}
	}

	// Use a simple JSON formatter for the error log file
	jsonFormatter := &formatter.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00", // ISO 8601
		EnableStackTrace: true,
	}

	return &SimpleFileHook{
		writer:    file,
		formatter: jsonFormatter,
		file:      file,
	}, nil
}

// Fire writes the log entry to the file.
func (h *SimpleFileHook) Fire(entry *core.LogEntry) error {
	if entry.Level < core.ERROR { // Only log ERROR level and above to the error file
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	buf := bytes.NewBuffer(nil) // Create a new buffer for each log entry
	if err := h.formatter.Format(buf, entry); err != nil {
		return &wrappedError{
			msg:   "file hook failed to format log entry",
			cause: err,
		}
	}

	buf.WriteByte('\n') // Add newline after each JSON entry

	if _, err := h.writer.Write(buf.Bytes()); err != nil {
		return &wrappedError{
			msg:   "file hook failed to write log entry",
			cause: err,
		}
	}
	return nil
}

// Close closes the underlying file writer.
func (h *SimpleFileHook) Close() error {
	if h.file != nil {
		return h.file.Close()
	}
	return nil
}
