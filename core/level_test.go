package core

import (
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{TRACE, "TRACE"},
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{NOTICE, "NOTICE"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{PANIC, "PANIC"},
		{(Level(99)), "UNKNOWN"}, // Test out-of-range level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() for %v = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLevel_Bytes(t *testing.T) {
	tests := []struct {
		level    Level
		expected []byte
	}{
		{TRACE, []byte("TRACE")},
		{DEBUG, []byte("DEBUG")},
		{INFO, []byte("INFO")},
		{NOTICE, []byte("NOTICE")},
		{WARN, []byte("WARN")},
		{ERROR, []byte("ERROR")},
		{FATAL, []byte("FATAL")},
		{PANIC, []byte("PANIC")},
		{(Level(99)), []byte("UNKNOWN")}, // Test out-of-range level
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			if got := tt.level.Bytes(); string(got) != string(tt.expected) {
				t.Errorf("Level.Bytes() for %v = %v, want %v", tt.level, string(got), string(tt.expected))
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Level
		expectError bool
	}{
		{"Valid TRACE", "TRACE", TRACE, false},
		{"Valid DEBUG lower", "debug", DEBUG, false},
		{"Valid INFO mixed", "Info", INFO, false},
		{"Valid NOTICE", "NOTICE", NOTICE, false},
		{"Valid WARN", "WARN", WARN, false},
		{"Valid ERROR", "ERROR", ERROR, false},
		{"Valid FATAL", "FATAL", FATAL, false},
		{"Valid PANIC", "PANIC", PANIC, false},
		{"Special WARNING", "WARNING", WARN, false},
		{"Invalid Level", "UNKNOWN", INFO, true}, // Should return INFO and an error
		{"Empty String", "", INFO, true},         // Should return INFO and an error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("ParseLevel() error status = %v, expectError %v", err != nil, tt.expectError)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseLevel() = %v, want %v for input %q", got, tt.expected, tt.input)
			}
		})
	}
}
