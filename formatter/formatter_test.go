package formatter

import (
	"bytes"
	"testing"

	"github.com/Lunar-Chipter/mire/core"
)

// TestFormatterInterface tests that the Formatter interface is properly defined
func TestFormatterInterface(t *testing.T) {
	// This test ensures that the Formatter interface can be implemented
	// We'll create a mock formatter to verify the interface
	var f Formatter
	if f != nil {
		t.Error("Formatter interface should be nil initially")
	}
	
	// Verify that we can assign implementations to the interface
	var textF Formatter = &TextFormatter{}
	var jsonF Formatter = &JSONFormatter{}
	var csvF Formatter = &CSVFormatter{}
	
	if textF == nil || jsonF == nil || csvF == nil {
		t.Error("Formatter implementations should not be nil")
	}
}

// TestAllFormatterInterface tests that the AllFormatter interface is properly defined
func TestAllFormatterInterface(t *testing.T) {
	// This test ensures that the AllFormatter interface can be implemented
	var af AllFormatter
	
	// Check that AllFormatter extends Formatter
	_ = Formatter(af)
	
	// Create a mock implementation that should satisfy AllFormatter
	// For now, just check the interface definition
}

// TestFormatterImplementation tests that formatters implement the Formatter interface
func TestFormatterImplementation(t *testing.T) {
	textFormatter := NewTextFormatter()
	jsonFormatter := NewJSONFormatter()
	csvFormatter := NewCSVFormatter()
	
	// Verify they implement the Formatter interface
	var _ Formatter = textFormatter
	var _ Formatter = jsonFormatter
	var _ Formatter = csvFormatter
	
}

// TestTextFormatterAsFormatter tests TextFormatter through the Formatter interface
func TestTextFormatterAsFormatter(t *testing.T) {
	f := NewTextFormatter()
	var formatter Formatter = f
	
	// Create a basic log entry for testing
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp // Set a valid timestamp
	
	buf := &bytes.Buffer{}
	err := formatter.Format(buf, entry)
	if err != nil {
		t.Errorf("TextFormatter.Format returned error: %v", err)
	}
	
	if buf.Len() == 0 {
		t.Error("TextFormatter.Format produced empty output")
	}
}

// TestJSONFormatterAsFormatter tests JSONFormatter through the Formatter interface
func TestJSONFormatterAsFormatter(t *testing.T) {
	f := NewJSONFormatter()
	var formatter Formatter = f
	
	// Create a basic log entry for testing
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp // Set a valid timestamp
	
	buf := &bytes.Buffer{}
	err := formatter.Format(buf, entry)
	if err != nil {
		t.Errorf("JSONFormatter.Format returned error: %v", err)
	}
	
	if buf.Len() == 0 {
		t.Error("JSONFormatter.Format produced empty output")
	}
}

// TestCSVFormatterAsFormatter tests CSVFormatter through the Formatter interface
func TestCSVFormatterAsFormatter(t *testing.T) {
	f := &CSVFormatter{IncludeHeader: false} // Disable header for simpler testing
	var formatter Formatter = f
	
	// Create a basic log entry for testing
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	entry.Timestamp = core.GetEntryFromPool().Timestamp // Set a valid timestamp
	
	buf := &bytes.Buffer{}
	err := formatter.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format returned error: %v", err)
	}
	
	if buf.Len() == 0 {
		t.Error("CSVFormatter.Format produced empty output")
	}
}

// TestResetColorBytes tests the ResetColorBytes constant
func TestResetColorBytes(t *testing.T) {
	expected := []byte("\033[0m")
	if string(ResetColorBytes) != string(expected) {
		t.Errorf("ResetColorBytes = %v, want %v", ResetColorBytes, expected)
	}
}

// TestMetaColorBytes tests the metaColorBytes constant
func TestMetaColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;245m") // Gray for meta info
	if string(metaColorBytes) != string(expected) {
		t.Errorf("metaColorBytes = %v, want %v", metaColorBytes, expected)
	}
}

// TestCallerColorBytes tests the callerColorBytes constant
func TestCallerColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;246m") // Gray for caller info
	if string(callerColorBytes) != string(expected) {
		t.Errorf("callerColorBytes = %v, want %v", callerColorBytes, expected)
	}
}

// TestDurationColorBytes tests the durationColorBytes constant
func TestDurationColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;155m") // Light green for duration
	if string(durationColorBytes) != string(expected) {
		t.Errorf("durationColorBytes = %v, want %v", durationColorBytes, expected)
	}
}

// TestTraceColorBytes tests the traceColorBytes constant
func TestTraceColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;141m") // Purple for trace info
	if string(traceColorBytes) != string(expected) {
		t.Errorf("traceColorBytes = %v, want %v", traceColorBytes, expected)
	}
}

// TestErrorColorBytes tests the errorColorBytes constant
func TestErrorColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;196m") // Bright red for errors
	if string(errorColorBytes) != string(expected) {
		t.Errorf("errorColorBytes = %v, want %v", errorColorBytes, expected)
	}
}

// TestStackTraceColorBytes tests the stackTraceColorBytes constant
func TestStackTraceColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;240m") // Dark gray for stack trace
	if string(stackTraceColorBytes) != string(expected) {
		t.Errorf("stackTraceColorBytes = %v, want %v", stackTraceColorBytes, expected)
	}
}

// TestFieldsWrapperColorBytes tests the fieldsWrapperColorBytes constant
func TestFieldsWrapperColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;243m") // Light gray for field wrappers
	if string(fieldsWrapperColorBytes) != string(expected) {
		t.Errorf("fieldsWrapperColorBytes = %v, want %v", fieldsWrapperColorBytes, expected)
	}
}

// TestFieldKeyColorBytes tests the fieldKeyColorBytes constant
func TestFieldKeyColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;228m") // Light yellow for field keys
	if string(fieldKeyColorBytes) != string(expected) {
		t.Errorf("fieldKeyColorBytes = %v, want %v", fieldKeyColorBytes, expected)
	}
}

// TestFieldValueColorBytes tests the fieldValueColorBytes constant
func TestFieldValueColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;159m") // Light cyan for field values
	if string(fieldValueColorBytes) != string(expected) {
		t.Errorf("fieldValueColorBytes = %v, want %v", fieldValueColorBytes, expected)
	}
}

// TestTagsColorBytes tests the tagsColorBytes constant
func TestTagsColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;135m") // Purple for tags
	if string(tagsColorBytes) != string(expected) {
		t.Errorf("tagsColorBytes = %v, want %v", tagsColorBytes, expected)
	}
}

// TestMetricsColorBytes tests the metricsColorBytes constant
func TestMetricsColorBytes(t *testing.T) {
	expected := []byte("\033[38;5;85m") // Green for metrics
	if string(metricsColorBytes) != string(expected) {
		t.Errorf("metricsColorBytes = %v, want %v", metricsColorBytes, expected)
	}
}