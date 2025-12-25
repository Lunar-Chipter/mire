package core

import (
	"bytes" // Added for bytes.EqualFold
	"strings"

	"github.com/Lunar-Chipter/mire/errors"
	"unsafe"
)

// ===============================
// LEVEL DEFINITION
// ===============================
// Level represents the severity level of a log entry
type Level int

const (
	// TRACE level for very detailed debugging information
	// TRACE level for very detailed debugging information
	TRACE Level = iota

	// DEBUG level for debugging information
	// DEBUG level for debugging information
	DEBUG

	// INFO level for general information messages
	// INFO level for general information messages
	INFO

	// at
	// at
	NOTICE

	// WARN level for warning messages
	// WARN level for warning messages
	WARN

	// ERROR level for error messages
	// ERROR level for error messages
	ERROR

	// FATAL level for critical errors that cause program termination
	// FATAL level for critical errors that cause program to stop
	FATAL

	// at
	// at
	PANIC
)

var (
	// LevelStrings contains the string representations of log levels
	LevelStrings = []string{
		"TRACE",
		"DEBUG",
		"INFO",
		"NOTICE",
		"WARN",
		"ERROR",
		"FATAL",
		"PANIC",
	}

	// LevelBytes contains the byte slice representations of log levels (zero-allocation for formatting)
	LevelBytes = func() [][]byte {
		b := make([][]byte, len(LevelStrings))
		for i, s := range LevelStrings {
			b[i] = []byte(s)
		}
		return b
	}()

	// LowerLevelStrings contains the lowercase string representations of log levels
	LowerLevelStrings []string = func() []string {
		lowers := make([]string, len(LevelStrings))
		for i, s := range LevelStrings {
			lowers[i] = strings.ToLower(s)
		}
		return lowers
	}()

	// LevelColors contains the ANSI color codes for each log level
	LevelColors = []string{
		"\033[38;5;246m", // Gray - TRACE
		"\033[36m",       // Cyan - DEBUG
		"\033[32m",       // Green - INFO
		"\033[38;5;220m", // Yellow - NOTICE
		"\033[33m",       // Orange - WARN
		"\033[31m",       // Red - ERROR
		"\033[38;5;198m", // Magenta - FATAL
		"\033[38;5;196m", // Bright Red - PANIC
	}

	// LevelColorBytes contains the ANSI color codes for each log level as byte slices
	LevelColorBytes = func() [][]byte {
		b := make([][]byte, len(LevelColors))
		for i, s := range LevelColors {
			b[i] = []byte(s)
		}
		return b
	}()

	// LevelBackgrounds contains the ANSI background color codes for each log level
	LevelBackgrounds = []string{
		"\033[48;5;238m", // Dark gray background
		"\033[48;5;236m",
		"\033[48;5;28m",
		"\033[48;5;94m",
		"\033[48;5;130m",
		"\033[48;5;88m",
		"\033[48;5;90m",
		"\033[48;5;52m",
	}

	// LevelBackgroundBytes contains the ANSI background color codes for each log level as byte slices
	LevelBackgroundBytes = func() [][]byte {
		b := make([][]byte, len(LevelBackgrounds))
		for i, s := range LevelBackgrounds {
			b[i] = []byte(s)
		}
		return b
	}()
)

// String returns the string representation of the level
func (l Level) String() string {
	if l >= TRACE && l <= PANIC {
		return LevelStrings[l]
	}
	return "UNKNOWN"
}

// Bytes returns the byte slice representation of the level (zero-allocation for formatting)
func (l Level) Bytes() []byte {
	if l >= TRACE && l <= PANIC {
		return LevelBytes[l]
	}
	return []byte("UNKNOWN") // Allocate in this rare case
}

// ToBytes returns the byte slice representation of the level for use in LogEntry
func (l Level) ToBytes() []byte {
	if l >= TRACE && l <= PANIC {
		return LevelBytes[l]
	}
	return []byte("UNKNOWN") // Allocate in this rare case
}

// s2b converts a string to a byte slice without memory allocation.
// WARNING: The returned byte slice shares memory with the string. It is read-only.
func s2b(s string) (b []byte) {
	bh := (*[3]int)(unsafe.Pointer(&b))
	sh := (*[2]int)(unsafe.Pointer(&s))
	bh[0] = sh[0]
	bh[1] = sh[1]
	bh[2] = sh[1]
	return b
}

// ParseLevel parses a level from string (case-insensitive)
func ParseLevel(levelStr string) (Level, error) {
	levelBytes := s2b(levelStr)

	for i, levelBytesSlice := range LevelBytes {
		if bytes.EqualFold(levelBytes, levelBytesSlice) {
			return Level(i), nil
		}
	}

	// Handle "WARNING" as alias for "WARN"
	warningBytes := []byte("WARNING")
	if bytes.EqualFold(levelBytes, warningBytes) {
		return WARN, nil
	}

	return INFO, errors.NewInvalidLevel(levelStr)
}
