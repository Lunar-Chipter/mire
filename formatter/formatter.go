package formatter

import (
	"bytes"
	"github.com/Lunar-Chipter/mire/core"
)

// Formatter interface defines how log entries are formatted
// Formatter interface defines how log entries are formatted
type Formatter interface {
	// Format formats a log entry into a byte slice
	// Format formats log entry into byte slice
	Format(buf *bytes.Buffer, entry *core.LogEntry) error
}

// at
type AllFormatter interface {
	Formatter
	// SetOptions allows setting formatter-specific options
	SetOptions(options interface{}) error
	// GetOptions returns the current formatter options
	GetOptions() interface{}
}
