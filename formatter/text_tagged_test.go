//go:build with_colors
// +build with_colors

package formatter

import (
	"bytes"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

func TestTextFormatter_WithColors(t *testing.T) {
	formatter := &TextFormatter{
		EnableColors: true,
	}

	entry := &core.LogEntry{
		Timestamp: time.Now(),
		Level:     core.INFO,
		Message:   []byte("test message"),
		Fields: map[string][]byte{
			"key": []byte("value"),
		},
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("\033[")) {
		t.Errorf("expected ANSI color codes, but none were found")
	}
}
