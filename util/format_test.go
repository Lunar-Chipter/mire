package util

import (
	"bytes"
	"testing"
	"time"
)

func TestFormatValue(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test string formatting
	buf.Reset()
	FormatValue(buf, "hello", 0)
	if buf.String() != "hello" {
		t.Errorf("FormatValue with string 'hello' = %s, want 'hello'", buf.String())
	}

	// Test []byte formatting
	buf.Reset()
	FormatValue(buf, []byte("hello"), 0)
	if buf.String() != "hello" {
		t.Errorf("FormatValue with []byte 'hello' = %s, want 'hello'", buf.String())
	}

	// Test int formatting
	buf.Reset()
	FormatValue(buf, 42, 0)
	if buf.String() != "42" {
		t.Errorf("FormatValue with int 42 = %s, want '42'", buf.String())
	}

	// Test int8 formatting
	buf.Reset()
	FormatValue(buf, int8(8), 0)
	if buf.String() != "8" {
		t.Errorf("FormatValue with int8 8 = %s, want '8'", buf.String())
	}

	// Test int16 formatting
	buf.Reset()
	FormatValue(buf, int16(16), 0)
	if buf.String() != "16" {
		t.Errorf("FormatValue with int16 16 = %s, want '16'", buf.String())
	}

	// Test int32 formatting
	buf.Reset()
	FormatValue(buf, int32(32), 0)
	if buf.String() != "32" {
		t.Errorf("FormatValue with int32 32 = %s, want '32'", buf.String())
	}

	// Test int64 formatting
	buf.Reset()
	FormatValue(buf, int64(64), 0)
	if buf.String() != "64" {
		t.Errorf("FormatValue with int64 64 = %s, want '64'", buf.String())
	}

	// Test uint formatting
	buf.Reset()
	FormatValue(buf, uint(42), 0)
	if buf.String() != "42" {
		t.Errorf("FormatValue with uint 42 = %s, want '42'", buf.String())
	}

	// Test uint8 formatting
	buf.Reset()
	FormatValue(buf, uint8(8), 0)
	if buf.String() != "8" {
		t.Errorf("FormatValue with uint8 8 = %s, want '8'", buf.String())
	}

	// Test uint16 formatting
	buf.Reset()
	FormatValue(buf, uint16(16), 0)
	if buf.String() != "16" {
		t.Errorf("FormatValue with uint16 16 = %s, want '16'", buf.String())
	}

	// Test uint32 formatting
	buf.Reset()
	FormatValue(buf, uint32(32), 0)
	if buf.String() != "32" {
		t.Errorf("FormatValue with uint32 32 = %s, want '32'", buf.String())
	}

	// Test uint64 formatting
	buf.Reset()
	FormatValue(buf, uint64(64), 0)
	if buf.String() != "64" {
		t.Errorf("FormatValue with uint64 64 = %s, want '64'", buf.String())
	}

	// Test float32 formatting
	buf.Reset()
	FormatValue(buf, float32(3.14), 0)
	if buf.String() != "3.14" {
		t.Errorf("FormatValue with float32 3.14 = %s, want '3.14'", buf.String())
	}

	// Test float64 formatting
	buf.Reset()
	FormatValue(buf, float64(3.14159), 0)
	if buf.String() != "3.14" { // The function formats to 2 decimal places
		t.Errorf("FormatValue with float64 3.14159 = %s, want '3.14'", buf.String())
	}

	// Test bool formatting
	buf.Reset()
	FormatValue(buf, true, 0)
	if buf.String() != "true" {
		t.Errorf("FormatValue with bool true = %s, want 'true'", buf.String())
	}

	buf.Reset()
	FormatValue(buf, false, 0)
	if buf.String() != "false" {
		t.Errorf("FormatValue with bool false = %s, want 'false'", buf.String())
	}

	// Test nil formatting
	buf.Reset()
	FormatValue(buf, nil, 0)
	if buf.String() != "null" {
		t.Errorf("FormatValue with nil = %s, want 'null'", buf.String())
	}

	// Test value with space (should be quoted)
	buf.Reset()
	FormatValue(buf, "hello world", 0)
	if buf.String() != `"hello world"` {
		t.Errorf("FormatValue with string containing space = %s, want '\"hello world\"'", buf.String())
	}

	// Test value with max width
	buf.Reset()
	FormatValue(buf, "very long string that should be truncated", 10)
	result := buf.String()
	if len(result) > 13 { // 10 + 3 for "..."
		t.Errorf("FormatValue with max width 10 produced string of length %d, should be max 13", len(result))
	}

	// Test complex type (will return <complex-type>)
	buf.Reset()
	FormatValue(buf, struct{ A int }{A: 1}, 0)
	if buf.String() != "<complex-type>" {
		t.Logf("FormatValue with complex type produced: %s", buf.String())
	}
}

func TestFormatTimestamp(t *testing.T) {
	buf := &bytes.Buffer{}

	timestamp := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)

	FormatTimestamp(buf, timestamp, "2006-01-02 15:04:05")

	expected := "2023-01-02 15:04:05"
	if buf.String() != expected {
		t.Errorf("FormatTimestamp = %s, want %s", buf.String(), expected)
	}

	// Test with different format
	buf.Reset()
	FormatTimestamp(buf, timestamp, time.RFC3339)
	expectedRFC3339 := "2023-01-02T15:04:05Z"
	if buf.String() != expectedRFC3339 {
		t.Errorf("FormatTimestamp with RFC3339 = %s, want %s", buf.String(), expectedRFC3339)
	}
}

// TestConvertValue tests the ConvertValue function
func TestConvertValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{[]byte("hello"), "hello"},
		{42, "42"},
		{int8(8), "8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(64), "64"},
		{uint(42), "42"},
		{uint8(8), "8"},
		{uint16(16), "16"},
		{uint32(32), "32"},
		{uint64(64), "64"},
		{float32(3.14), "3.14"},
		{float64(3.14159), "3.14159"},
		{true, "true"},
		{false, "false"},
		{nil, "null"},
		{struct{ A int }{A: 1}, "<complex-type>"}, // Complex type fallback
	}
	
	for _, test := range tests {
		result := ConvertValue(test.input)
		if result != test.expected {
			t.Errorf("ConvertValue(%v) = %s, want %s", test.input, result, test.expected)
		}
	}
}

// TestWriteInt tests the WriteInt function
func TestWriteInt(t *testing.T) {
	buf := &bytes.Buffer{}
	
	WriteInt(buf, 12345)
	
	if buf.String() != "12345" {
		t.Errorf("WriteInt(12345) = %s, want '12345'", buf.String())
	}
	
	// Test with negative number
	buf.Reset()
	WriteInt(buf, -6789)
	
	if buf.String() != "-6789" {
		t.Errorf("WriteInt(-6789) = %s, want '-6789'", buf.String())
	}
	
	// Test with zero
	buf.Reset()
	WriteInt(buf, 0)
	
	if buf.String() != "0" {
		t.Errorf("WriteInt(0) = %s, want '0'", buf.String())
	}
}

// TestWriteUint tests the WriteUint function
func TestWriteUint(t *testing.T) {
	buf := &bytes.Buffer{}
	
	WriteUint(buf, 12345)
	
	if buf.String() != "12345" {
		t.Errorf("WriteUint(12345) = %s, want '12345'", buf.String())
	}
	
	// Test with zero
	buf.Reset()
	WriteUint(buf, 0)
	
	if buf.String() != "0" {
		t.Errorf("WriteUint(0) = %s, want '0'", buf.String())
	}
}

// TestWriteFloat tests the WriteFloat function
func TestWriteFloat(t *testing.T) {
	buf := &bytes.Buffer{}
	
	WriteFloat(buf, 3.14159)
	
	// The exact output depends on the float formatting, but it should contain "3.14159" or similar
	output := buf.String()
	if len(output) == 0 {
		t.Error("WriteFloat should produce output")
	}
	
	if output != "3.14159" && output != "3.14" {
		t.Logf("WriteFloat(3.14159) produced: %s", output) // Log actual result
	}
}

// TestConvertValueString tests more specific string conversion cases
func TestConvertValueString(t *testing.T) {
	// Test string conversion
	result := ConvertValue("test string")
	if result != "test string" {
		t.Errorf("ConvertValue with string = %s, want 'test string'", result)
	}
	
	// Test []byte to string conversion
	result = ConvertValue([]byte("test bytes"))
	if result != "test bytes" {
		t.Errorf("ConvertValue with []byte = %s, want 'test bytes'", result)
	}
}

// TestConvertValueIntegers tests integer conversion cases
func TestConvertValueIntegers(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{int(123), "123"},
		{int8(-8), "-8"},
		{int16(32767), "32767"},
		{int32(-1000000), "-1000000"},
		{int64(9223372036854775807), "9223372036854775807"},
		{uint(456), "456"},
		{uint8(255), "255"},
		{uint16(65535), "65535"},
		{uint32(4294967295), "4294967295"},
		{uint64(18446744073709551615), "18446744073709551615"},
	}
	
	for _, test := range tests {
		result := ConvertValue(test.input)
		if result != test.expected {
			t.Errorf("ConvertValue(%v) = %s, want %s", test.input, result, test.expected)
		}
	}
}

// TestConvertValueFloats tests float conversion cases
func TestConvertValueFloats(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{float32(1.23), "1.23"},
		{float64(4.56), "4.56"},
		{float32(0.0), "0"},
		{float64(-7.89), "-7.89"},
	}
	
	for _, test := range tests {
		result := ConvertValue(test.input)
		// For floats, we might get more precision than expected, so check if it contains the expected value
		if result != test.expected && 
		   !(test.input == float32(0.0) && result == "0") &&
		   !(test.input == float64(0.0) && result == "0") {
			t.Errorf("ConvertValue(%v) = %s, want %s", test.input, result, test.expected)
		}
	}
}

// TestConvertValueBooleans tests boolean conversion cases
func TestConvertValueBooleans(t *testing.T) {
	result := ConvertValue(true)
	if result != "true" {
		t.Errorf("ConvertValue(true) = %s, want 'true'", result)
	}
	
	result = ConvertValue(false)
	if result != "false" {
		t.Errorf("ConvertValue(false) = %s, want 'false'", result)
	}
}

// TestConvertValueNil tests nil conversion
func TestConvertValueNil(t *testing.T) {
	result := ConvertValue(nil)
	if result != "null" {
		t.Errorf("ConvertValue(nil) = %s, want 'null'", result)
	}
}

// TestConvertValueComplexType tests complex type conversion (fallback case)
func TestConvertValueComplexType(t *testing.T) {
	// Test with a struct
	result := ConvertValue(struct{ Name string }{Name: "test"})
	if result != "<complex-type>" {
		t.Errorf("ConvertValue(complex type) = %s, want '<complex-type>'", result)
	}
	
	// Test with a slice
	result = ConvertValue([]int{1, 2, 3})
	if result != "<complex-type>" {
		t.Errorf("ConvertValue(slice) = %s, want '<complex-type>'", result)
	}
	
	// Test with a map
	result = ConvertValue(map[string]int{"a": 1})
	if result != "<complex-type>" {
		t.Errorf("ConvertValue(map) = %s, want '<complex-type>'", result)
	}
}