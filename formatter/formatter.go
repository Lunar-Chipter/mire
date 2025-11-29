package formatter

import (
	"bytes"
	"github.com/Lunar-Chipter/mire/core"
)

// Formatter interface defines how log entries are formatted
// Interface Formatter mendefinisikan bagaimana entri log diformat
type Formatter interface {
	// Format formats a log entry into a byte slice
	// Format memformat entri log menjadi slice byte
	Format(buf *bytes.Buffer, entry *core.LogEntry) error
}

// AllFormatter interface defines additional methods for formatters that support various options
type AllFormatter interface {
	Formatter
	// SetOptions allows setting formatter-specific options
	SetOptions(options interface{}) error
	// GetOptions returns the current formatter options
	GetOptions() interface{}
}
