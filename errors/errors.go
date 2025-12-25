package errors

import (
	"bytes"
	"sync"
)

// InvalidLevel is a custom error type for invalid log levels
type InvalidLevel struct {
	level string
	buf   *bytes.Buffer
}

// invalidLevelPool pools InvalidLevel objects
var invalidLevelPool = sync.Pool{
	New: func() interface{} {
		return &InvalidLevel{
			buf: new(bytes.Buffer),
		}
	},
}

// NewInvalidLevel gets a pooled InvalidLevel
func NewInvalidLevel(level string) *InvalidLevel {
	err := invalidLevelPool.Get().(*InvalidLevel)
	err.level = level
	err.buf.Reset()
	return err
}

// PutInvalidLevel returns InvalidLevel to pool
func PutInvalidLevel(err *InvalidLevel) {
	invalidLevelPool.Put(err)
}

// AppendError implements the ErrAppend interface for InvalidLevel.
func (e *InvalidLevel) AppendError(buf *bytes.Buffer) {
	buf.WriteString("invalid log level: ")
	buf.WriteString(e.level)
}

// Error returns error message
func (e *InvalidLevel) Error() string {
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
