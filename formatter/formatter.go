package formatter

import (
	"bytes"
	"mire/core"
)

// Formatter interface defines how log entries are formatted
// Interface Formatter mendefinisikan bagaimana entri log diformat
type Formatter interface {
	// Format formats a log entry into a byte slice
	// Format memformat entri log menjadi slice byte
	Format(buf *bytes.Buffer, entry *core.LogEntry) error
}
