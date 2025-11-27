package errors

import (
	"bytes"
	"sync"
)

// InvalidLogLevelError is a custom error type for invalid log levels
type InvalidLogLevelError struct {
	level string
	buf   *bytes.Buffer // Pooled buffer for error message
}

// invalidLogLevelErrorPool pools InvalidLogLevelError objects
var invalidLogLevelErrorPool = sync.Pool{
	New: func() interface{} {
		return &InvalidLogLevelError{
			buf: new(bytes.Buffer),
		}
	},
}

// NewInvalidLogLevelError gets a pooled InvalidLogLevelError
func NewInvalidLogLevelError(level string) *InvalidLogLevelError {
	err := invalidLogLevelErrorPool.Get().(*InvalidLogLevelError)
	err.level = level
	err.buf.Reset() // Reset the buffer
	return err
}

// PutInvalidLogLevelError returns the InvalidLogLevelError to the pool
func PutInvalidLogLevelError(err *InvalidLogLevelError) {
	invalidLogLevelErrorPool.Put(err)
}

// AppendError implements the ErrorAppender interface for InvalidLogLevelError.
func (e *InvalidLogLevelError) AppendError(buf *bytes.Buffer) {
	buf.WriteString("invalid log level: ")
	buf.WriteString(e.level)
}

// Error returns the error message (for standard error interface compatibility)
func (e *InvalidLogLevelError) Error() string {
	// Re-use internal buffer to build string, but this still allocates a string.
	// This method is primarily for compatibility with the standard error interface.
	e.buf.Reset()
	e.AppendError(e.buf)
	return e.buf.String()
}

// customError is a simple error implementation
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

var ErrAsyncBufferFull = &customError{msg: "async log channel full"}
