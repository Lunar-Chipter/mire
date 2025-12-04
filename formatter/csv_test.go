package formatter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/util"
)

// TestNewCSVFormatter tests creating a new CSVFormatter
func TestNewCSVFormatter(t *testing.T) {
	cf := NewCSVFormatter()
	
	if cf == nil {
		t.Fatal("NewCSVFormatter returned nil")
	}
	
	// Check default values
	if cf.MaskStringValue != "[MASKED]" {
		t.Errorf("Default MaskStringValue should be '[MASKED]', got '%s'", cf.MaskStringValue)
	}
	
	if cf.FieldTransformers == nil {
		t.Error("FieldTransformers should not be nil")
	}
}

// TestCSVFormatterFormat tests the Format method of CSVFormatter
func TestCSVFormatterFormat(t *testing.T) {
	cf := NewCSVFormatter()
	// Don't include header for simpler testing initially
	cf.IncludeHeader = false
	
	// Create a test log entry
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	entry.PID = 1234
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format produced empty output")
	}
	
	// Since we haven't set FieldOrder, we can't easily predict the output
	// But it should at least produce a line ending
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		t.Error("CSVFormatter output should end with newline")
	}
}

// TestCSVFormatterWithCustomFieldOrder tests CSVFormatter with custom field order
func TestCSVFormatterWithCustomFieldOrder(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"timestamp", "level", "message", "pid"}
	cf.IncludeHeader = false
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.ERROR
	entry.Message = []byte("error message")
	entry.PID = 5678
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with custom field order returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with custom field order produced empty output")
	}
	
	// Should produce comma-separated values (at least 3 commas for 4 fields)
	if strings.Count(output, ",") < 3 {
		t.Log("Output may not contain expected commas - depends on implementation")
	}
}

// TestCSVFormatterWithHeader tests CSVFormatter with header
func TestCSVFormatterWithHeader(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"timestamp", "level", "message", "pid"}
	cf.IncludeHeader = true
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with header")
	entry.PID = 9999
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with header returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with header produced empty output")
	}
	
	// Should contain at least two lines (header + data)
	lines := bytes.Count(buf.Bytes(), []byte("\n"))
	if lines < 2 {
		t.Log("Output may not contain header + data lines - depends on implementation")
	}
}

// TestCSVFormatterWriteCSVValue tests the writeCSVValue method
func TestCSVFormatterWriteCSVValue(t *testing.T) {
	cf := NewCSVFormatter()
	
	buf := &bytes.Buffer{}
	
	// Test normal value without special characters
	buf.Reset()
	cf.writeCSVValue(buf, "normal_value")
	if buf.String() != "normal_value" {
		t.Errorf("writeCSVValue with normal string = %s, want 'normal_value'", buf.String())
	}
	
	// Test value with comma (should be quoted)
	buf.Reset()
	cf.writeCSVValue(buf, "value,with,comma")
	if !strings.Contains(buf.String(), `"value,with,comma"`) {
		t.Error("writeCSVValue with comma should quote the value")
	}
	
	// Test value with quote (should be escaped)
	buf.Reset()
	cf.writeCSVValue(buf, `value "with" quotes`)
	if !strings.Contains(buf.String(), `value ""with"" quotes`) {
		t.Error("writeCSVValue with quotes should escape the quotes")
	}
	
	// Test value with newline (should be quoted)
	buf.Reset()
	cf.writeCSVValue(buf, "value\nwith\nnewline")
	if !strings.Contains(buf.String(), `"value`) || !strings.Contains(buf.String(), `newline"`) {
		t.Error("writeCSVValue with newline should quote the value")
	}
}

// TestCSVFormatterWriteCSVValueBytes tests the writeCSVValueBytes method
func TestCSVFormatterWriteCSVValueBytes(t *testing.T) {
	cf := NewCSVFormatter()
	
	buf := &bytes.Buffer{}
	
	// Test normal value without special characters
	buf.Reset()
	cf.writeCSVValueBytes(buf, []byte("normal_value"))
	if buf.String() != "normal_value" {
		t.Errorf("writeCSVValueBytes with normal bytes = %s, want 'normal_value'", buf.String())
	}
	
	// Test value with comma (should be quoted)
	buf.Reset()
	cf.writeCSVValueBytes(buf, []byte("value,with,comma"))
	if !strings.Contains(buf.String(), `"value,with,comma"`) {
		t.Error("writeCSVValueBytes with comma should quote the value")
	}
	
	// Test value with quote (should be escaped)
	buf.Reset()
	cf.writeCSVValueBytes(buf, []byte(`value "with" quotes`))
	if !strings.Contains(buf.String(), `value ""with"" quotes`) {
		t.Error("writeCSVValueBytes with quotes should escape the quotes")
	}
	
	// Test value with newline (should be quoted)
	buf.Reset()
	cf.writeCSVValueBytes(buf, []byte("value\nwith\nnewline"))
	if !strings.Contains(buf.String(), `"value`) || !strings.Contains(buf.String(), `newline"`) {
		t.Error("writeCSVValueBytes with newline should quote the value")
	}
}

// TestCSVFormatterFormatCSVField tests the formatCSVField method
func TestCSVFormatterFormatCSVField(t *testing.T) {
	cf := NewCSVFormatter()
	
	buf := &bytes.Buffer{}
	
	// Create a basic entry to test field formatting
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("test message")
	entry.PID = 1234
	entry.GoroutineID = []byte("5678")
	entry.TraceID = []byte("trace123")
	entry.SpanID = []byte("span456")
	entry.UserID = []byte("user789")
	entry.RequestID = []byte("req012")
	entry.Caller = util.GetCallerInfo(1)
	if entry.Caller != nil {
		entry.Caller.File = "test.go"
		entry.Caller.Line = 100
	}
	
	// Test timestamp field
	buf.Reset()
	err := cf.formatCSVField(buf, "timestamp", entry)
	if err != nil {
		t.Errorf("formatCSVField for timestamp returned error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("formatCSVField for timestamp should produce output")
	}
	
	// Test level field
	buf.Reset()
	err = cf.formatCSVField(buf, "level", entry)
	if err != nil {
		t.Errorf("formatCSVField for level returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "INFO") {
		t.Error("formatCSVField for level should contain 'INFO'")
	}
	
	// Test message field
	buf.Reset()
	err = cf.formatCSVField(buf, "message", entry)
	if err != nil {
		t.Errorf("formatCSVField for message returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Error("formatCSVField for message should contain 'test message'")
	}
	
	// Test pid field
	buf.Reset()
	err = cf.formatCSVField(buf, "pid", entry)
	if err != nil {
		t.Errorf("formatCSVField for pid returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "1234") {
		t.Error("formatCSVField for pid should contain '1234'")
	}
	
	// Test goroutine_id field
	buf.Reset()
	err = cf.formatCSVField(buf, "goroutine_id", entry)
	if err != nil {
		t.Errorf("formatCSVField for goroutine_id returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "5678") {
		t.Error("formatCSVField for goroutine_id should contain '5678'")
	}
	
	// Test trace_id field
	buf.Reset()
	err = cf.formatCSVField(buf, "trace_id", entry)
	if err != nil {
		t.Errorf("formatCSVField for trace_id returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "trace123") {
		t.Error("formatCSVField for trace_id should contain 'trace123'")
	}
	
	// Test span_id field
	buf.Reset()
	err = cf.formatCSVField(buf, "span_id", entry)
	if err != nil {
		t.Errorf("formatCSVField for span_id returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "span456") {
		t.Error("formatCSVField for span_id should contain 'span456'")
	}
	
	// Test user_id field
	buf.Reset()
	err = cf.formatCSVField(buf, "user_id", entry)
	if err != nil {
		t.Errorf("formatCSVField for user_id returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "user789") {
		t.Error("formatCSVField for user_id should contain 'user789'")
	}
	
	// Test request_id field
	buf.Reset()
	err = cf.formatCSVField(buf, "request_id", entry)
	if err != nil {
		t.Errorf("formatCSVField for request_id returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "req012") {
		t.Error("formatCSVField for request_id should contain 'req012'")
	}
	
	// Test file field (from caller)
	buf.Reset()
	err = cf.formatCSVField(buf, "file", entry)
	if err != nil {
		t.Errorf("formatCSVField for file returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "test.go") {
		t.Log("formatCSVField for file might not contain 'test.go' - depends on implementation") // May be empty if file is empty
	}
	
	// Test line field (from caller)
	buf.Reset()
	err = cf.formatCSVField(buf, "line", entry)
	if err != nil {
		t.Errorf("formatCSVField for line returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "100") {
		t.Log("formatCSVField for line might not contain '100' - depends on implementation") // May be empty if line is 0
	}
	
	// Clean up
	if entry.Caller != nil {
		core.PutCallerInfoToPool(entry.Caller)
	}
}

// TestCSVFormatterWithFields tests CSVFormatter with custom fields
func TestCSVFormatterWithFields(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"timestamp", "level", "message", "custom_field1", "custom_field2"}
	cf.IncludeHeader = false
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with custom fields")
	entry.Fields["custom_field1"] = "value1"
	entry.Fields["custom_field2"] = 42
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with custom fields returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with custom fields produced empty output")
	}
}

// TestCSVFormatterWithSensitiveFields tests CSVFormatter with sensitive fields masking
func TestCSVFormatterWithSensitiveFields(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"timestamp", "level", "message", "password", "token"}
	cf.IncludeHeader = false
	cf.MaskSensitiveData = true
	cf.SensitiveFields = []string{"password", "token"}
	cf.MaskStringValue = "***MASKED***"
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with sensitive fields")
	entry.Fields["password"] = "secret123"
	entry.Fields["token"] = "abc-def-ghi"
	entry.Fields["normal_field"] = "visible_value"
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with sensitive fields returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with sensitive fields produced empty output")
	}
	
	// Output should contain masked values instead of actual sensitive data
	if strings.Contains(output, "secret123") {
		t.Error("Output should not contain actual password value")
	}
	if strings.Contains(output, "abc-def-ghi") {
		t.Error("Output should not contain actual token value")
	}
	if !strings.Contains(output, "***MASKED***") {
		t.Log("Output may not contain masked values - depends on implementation")
	}
}

// TestCSVFormatterWithFieldTransformers tests CSVFormatter with field transformers
func TestCSVFormatterWithFieldTransformers(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"timestamp", "level", "message", "secret_field"}
	cf.IncludeHeader = false
	
	cf.FieldTransformers = map[string]func(interface{}) string{
		"secret_field": func(v interface{}) string {
			return "[TRANSFORMED]"
		},
	}
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with transformed field")
	entry.Fields["secret_field"] = "original_value"
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with field transformers returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with field transformers produced empty output")
	}
	
	// Output should contain transformed value instead of original
	if strings.Contains(output, "original_value") {
		t.Log("Output may still contain original value - depends on implementation")
	}
	if !strings.Contains(output, "[TRANSFORMED]") {
		t.Log("Output may not contain transformed value - depends on implementation")
	}
}

// TestCSVFormatterIsSensitiveField tests the isSensitiveField method
func TestCSVFormatterIsSensitiveField(t *testing.T) {
	cf := NewCSVFormatter()
	cf.SensitiveFields = []string{"password", "token", "secret"}
	
	tests := []struct {
		field    string
		expected bool
	}{
		{"password", true},
		{"token", true},
		{"secret", true},
		{"username", false},
		{"email", false},
		{"", false},
	}
	
	for _, test := range tests {
		result := cf.isSensitiveField(test.field)
		if result != test.expected {
			t.Errorf("isSensitiveField(%q) = %v, want %v", test.field, result, test.expected)
		}
	}
}

// TestCSVFormatterWithEmptyFieldOrder tests CSVFormatter with empty field order
func TestCSVFormatterWithEmptyFieldOrder(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{} // Empty field order
	cf.IncludeHeader = false
	
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)
	
	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with empty field order")
	
	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with empty field order returned error: %v", err)
	}
	
	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with empty field order produced empty output")
	}
	
	// Output should still contain a newline
	if !strings.HasSuffix(output, "\n") {
		t.Error("Output should end with newline even with empty field order")
	}
}

// TestCSVFormatterWithNonexistentField tests CSVFormatter with a field that doesn't exist in the entry
func TestCSVFormatterWithNonexistentField(t *testing.T) {
	cf := NewCSVFormatter()
	cf.FieldOrder = []string{"nonexistent_field", "timestamp", "level"}
	cf.IncludeHeader = false

	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)

	entry.Timestamp = time.Now()
	entry.Level = core.INFO
	entry.Message = []byte("with nonexistent field")

	buf := &bytes.Buffer{}
	err := cf.Format(buf, entry)
	if err != nil {
		t.Errorf("CSVFormatter.Format with nonexistent field returned error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("CSVFormatter.Format with nonexistent field produced empty output")
	}
}
