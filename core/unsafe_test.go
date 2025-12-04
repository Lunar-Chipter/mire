package core

import (
	"testing"
)

// TestStringToBytes tests the StringToBytes function
func TestStringToBytes(t *testing.T) {
	testString := "hello world"
	
	// Convert string to bytes
	result := StringToBytes(testString)
	
	// Verify the result has the correct content
	if string(result) != testString {
		t.Errorf("StringToBytes result doesn't match original string. Got '%s', want '%s'", string(result), testString)
	}
	
	// Verify the length is correct
	if len(result) != len(testString) {
		t.Errorf("StringToBytes result length %d doesn't match original length %d", len(result), len(testString))
	}
	
	// Verify the capacity is at least the length
	if cap(result) < len(result) {
		t.Error("StringToBytes result capacity is less than its length")
	}
	
	// Test with empty string
	emptyResult := StringToBytes("")
	if len(emptyResult) != 0 {
		t.Error("StringToBytes with empty string should return empty slice")
	}
	
	// Test with a longer string
	longString := "This is a longer test string to ensure the conversion works with various lengths"
	longResult := StringToBytes(longString)
	if string(longResult) != longString {
		t.Errorf("StringToBytes with long string failed. Got '%s', want '%s'", string(longResult), longString)
	}
}

// TestBytesToString tests the BytesToString function
func TestBytesToString(t *testing.T) {
	testBytes := []byte("hello world")
	
	// Convert bytes to string
	result := BytesToString(testBytes)
	
	// Verify the result matches the original
	if result != string(testBytes) {
		t.Errorf("BytesToString result doesn't match original. Got '%s', want '%s'", result, string(testBytes))
	}
	
	// Test with empty byte slice
	emptyBytes := []byte{}
	emptyResult := BytesToString(emptyBytes)
	if emptyResult != "" {
		t.Error("BytesToString with empty slice should return empty string")
	}
	
	// Test with a longer byte slice
	longBytes := []byte("This is a longer test byte slice to ensure the conversion works with various lengths")
	longResult := BytesToString(longBytes)
	if longResult != string(longBytes) {
		t.Errorf("BytesToString with long byte slice failed. Got '%s', want '%s'", longResult, string(longBytes))
	}
	
	// Test with byte slice containing special characters
	specialBytes := []byte("Special chars: \n \t \r \" ' & < > 日本語 Ελληνικά")
	specialResult := BytesToString(specialBytes)
	if specialResult != string(specialBytes) {
		t.Errorf("BytesToString with special chars failed")
	}
}

// TestStringToBytesAndBack tests the round-trip conversion
func TestStringToBytesAndBack(t *testing.T) {
	original := "test round-trip conversion"
	
	// String -> Bytes
	bytes := StringToBytes(original)
	
	// Bytes -> String
	result := BytesToString(bytes)
	
	// Should get back the original string
	if result != original {
		t.Errorf("Round-trip conversion failed. Original: '%s', Result: '%s'", original, result)
	}
}

// TestStringToBytesMemorySharing tests if the conversion shares memory
func TestStringToBytesMemorySharing(t *testing.T) {
	// This test is more complex because we can't safely modify the result
	// of StringToBytes without undefined behavior.
	// The purpose is to document the unsafe nature of the conversion.
	
	originalString := "test memory sharing"
	
	// Get the converted bytes
	convertedBytes := StringToBytes(originalString)
	
	// Verify they have the same content
	if string(convertedBytes) != originalString {
		t.Error("Converted bytes don't match original string")
	}
	
	// The conversion shares memory, which means modifying the result
	// is unsafe and could affect the original string or cause corruption.
	// We can't safely test this without causing undefined behavior.
	// The test ensures that the function doesn't crash or panic.
}

// TestBytesToStringMemorySharing tests if the conversion shares memory
func TestBytesToStringMemorySharing(t *testing.T) {
	// Similar to the above, the BytesToString conversion shares memory
	originalBytes := []byte("test memory sharing")
	
	// Convert to string
	convertedString := BytesToString(originalBytes)
	
	// Verify they match
	if convertedString != string(originalBytes) {
		t.Error("Converted string doesn't match original bytes")
	}
	
	// Again, we can't safely modify the original byte slice to see if the string changes
	// because it would be unsafe behavior.
}

// TestStringToBytesWithVariousInputs tests StringToBytes with various input strings
func TestStringToBytesWithVariousInputs(t *testing.T) {
	testCases := []string{
		"",                    // Empty string
		"a",                   // Single character
		"12345",              // Numbers
		"Hello, 世界",          // Unicode
		"line1\nline2",       // Newline
		"tab\there",          // Tab
		"quote\"test",        // Quote
		"special chars: !@#$%^&*()", // Special characters
		"κόσμε",              // Greek characters
		"Здравствуй",         // Cyrillic characters
		"مرحبا",               // Arabic characters
	}
	
	for i, testCase := range testCases {
		t.Run(string(rune(i)), func(t *testing.T) {
			result := StringToBytes(testCase)
			
			// Verify conversion works
			if string(result) != testCase {
				t.Errorf("Test case %d failed. Input: '%s', Output: '%s'", i, testCase, string(result))
			}
			
			// Verify length matches
			if len(result) != len(testCase) {
				t.Errorf("Length mismatch for test case %d. Expected: %d, Got: %d", 
					i, len(testCase), len(result))
			}
		})
	}
}

// TestBytesToStringWithVariousInputs tests BytesToString with various input byte slices
func TestBytesToStringWithVariousInputs(t *testing.T) {
	testCases := [][]byte{
		{},                           // Empty slice
		{'a'},                       // Single byte
		{'1', '2', '3', '4', '5'},  // Number bytes
		[]byte("Hello, 世界"),        // Unicode bytes
		[]byte("line1\nline2"),      // Newline bytes
		[]byte("tab\there"),         // Tab bytes
		[]byte("quote\"test"),       // Quote bytes
		[]byte("special chars: !@#$%^&*()"), // Special character bytes
		[]byte("κόσμε"),             // Greek bytes
		[]byte("Здравствуй"),        // Cyrillic bytes
		[]byte("مرحبا"),             // Arabic bytes
	}
	
	for i, testCase := range testCases {
		t.Run(string(rune(i)), func(t *testing.T) {
			result := BytesToString(testCase)
			
			// Verify conversion works
			if result != string(testCase) {
				t.Errorf("Test case %d failed. Input: '%s', Output: '%s'", i, string(testCase), result)
			}
		})
	}
}

// TestUnsafeSize tests that the unsafe operations work within expected size limits
func TestUnsafeSize(t *testing.T) {
	// Create a string and convert to bytes
	testStr := "test string"
	bytes := StringToBytes(testStr)
	
	// Verify that the unsafe operation doesn't create invalid slice
	if len(bytes) != len(testStr) {
		t.Error("Length mismatch in unsafe conversion")
	}
	
	if cap(bytes) < len(bytes) {
		t.Error("Capacity is less than length after unsafe conversion")
	}
	
	// Verify the string conversion back works
	str := BytesToString(bytes)
	if str != testStr {
		t.Error("Round-trip unsafe conversion failed")
	}
}

// TestStringToBytesLargeString tests with a large string
func TestStringToBytesLargeString(t *testing.T) {
	// Create a large string
	largeString := make([]byte, 10000)
	for i := range largeString {
		largeString[i] = byte('A' + (i % 26)) // Cycle through A-Z
	}
	testStr := string(largeString)
	
	// Convert to bytes
	result := StringToBytes(testStr)
	
	// Verify the conversion maintains the data
	if string(result) != testStr {
		t.Error("Large string conversion failed")
	}
	
	if len(result) != len(testStr) {
		t.Errorf("Length mismatch for large string: expected %d, got %d", len(testStr), len(result))
	}
}

// TestBytesToStringLargeBytes tests with large byte slice
func TestBytesToStringLargeBytes(t *testing.T) {
	// Create a large byte slice
	largeBytes := make([]byte, 10000)
	for i := range largeBytes {
		largeBytes[i] = byte('A' + (i % 26)) // Cycle through A-Z
	}
	
	// Convert to string
	result := BytesToString(largeBytes)
	
	// Verify the conversion maintains the data
	if result != string(largeBytes) {
		t.Error("Large byte slice conversion failed")
	}
	
	if len(result) != len(string(largeBytes)) {
		t.Errorf("Length mismatch for large byte slice: expected %d, got %d", len(string(largeBytes)), len(result))
	}
}