package core

import (
	"testing"
)

func TestLevelString(t *testing.T) {
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
		{-1, "UNKNOWN"},
		{8, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLevelBytes(t *testing.T) {
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
		{-1, []byte("UNKNOWN")},
		{8, []byte("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			got := tt.level.Bytes()
			if len(got) != len(tt.expected) {
				t.Errorf("Level.Bytes() length = %v, want %v", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("Level.Bytes()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestLevelToBytes(t *testing.T) {
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
		{-1, []byte("UNKNOWN")},
		{8, []byte("UNKNOWN")},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			got := tt.level.ToBytes()
			if len(got) != len(tt.expected) {
				t.Errorf("Level.ToBytes() length = %v, want %v", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("Level.ToBytes()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		hasError bool
	}{
		{"trace", TRACE, false},
		{"DEBUG", DEBUG, false},
		{"info", INFO, false},
		{"Notice", NOTICE, false},
		{"warn", WARN, false},
		{"WARNING", WARN, false},
		{"error", ERROR, false},
		{"fatal", FATAL, false},
		{"panic", PANIC, false},
		{"invalid", INFO, true}, // Should return INFO as default with error
		{"", INFO, true},        // Should return INFO as default with error
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("ParseLevel(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLevel(%q) expected no error, got %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}