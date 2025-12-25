package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/config"
)

// TestNewRotator tests creating a new Rotator
func TestNewRotator(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.log")

	// Create a rotation config
	rotationConfig := &config.RotationConfig{
		MaxSize:         1024, // 1KB
		MaxAge:          24 * time.Hour,
		MaxBackups:      5,
		LocalTime:       true,
		Compress:        false,
		RotateInterval:    time.Hour,
		FilenamePattern: "2006-01-02",
	}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	if rotatingWriter == nil {
		t.Fatal("NewRotator returned nil")
	}
	defer rotatingWriter.Close()

	// Verify the file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", tempFile)
	}

	// Verify the file field is set correctly
	if rotatingWriter.file == nil {
		t.Error("Rotator file field is nil")
	}
}

// TestNewRotatorError tests error case when file can't be created
func TestNewRotatorError(t *testing.T) {
	// Try to create a writer with an invalid path
	rotationConfig := &config.RotationConfig{}

	rotatingWriter, err := NewRotator("/invalid/path/that/does/not/exist/file.log", rotationConfig)

	if err == nil {
		t.Error("NewRotator should have failed with invalid path")
		if rotatingWriter != nil {
			rotatingWriter.Close()
		}
		return
	}

	if rotatingWriter != nil {
		t.Error("NewRotator should have returned nil on error")
		rotatingWriter.Close()
	}
}

// TestRotatorWrite tests writing to the Rotator
func TestRotatorWrite(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_write.log")

	rotationConfig := &config.RotationConfig{
		MaxSize: 1024, // Small size to trigger rotation easily
	}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	defer rotatingWriter.Close()

	// Write data
	data := []byte("This is a test log entry\n")
	n, err := rotatingWriter.Write(data)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, expected %d", n, len(data))
	}

	// Check that the file contains the data
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	} else if string(content) != string(data) {
		t.Errorf("File content '%s' doesn't match written data '%s'", string(content), string(data))
	}
}

// TestRotatorMultipleWrites tests multiple writes to the Rotator
func TestRotatorMultipleWrites(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_multiple.log")

	rotationConfig := &config.RotationConfig{
		MaxSize: 2048, // Larger size for multiple writes
	}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	defer rotatingWriter.Close()

	// Write multiple pieces of data
	expectedContent := ""
	for i := 0; i < 10; i++ {
		data := []byte("Log entry " + string(rune(i+'0')) + "\n")
		n, err := rotatingWriter.Write(data)
		if err != nil {
			t.Errorf("Write %d returned error: %v", i, err)
		}
		if n != len(data) {
			t.Errorf("Write %d returned %d, expected %d", i, n, len(data))
		}

		expectedContent += string(data)
	}

	// Check that the file contains all the data
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	} else if string(content) != expectedContent {
		t.Errorf("File content doesn't match expected. Got: %s, Expected: %s", string(content), expectedContent)
	}
}

// TestRotatorClose tests closing the Rotator
func TestRotatorClose(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_close.log")

	rotationConfig := &config.RotationConfig{}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}

	// Write some data
	_, _ = rotatingWriter.Write([]byte("Before close\n"))

	// Close the writer
	err = rotatingWriter.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Try to close again - should not cause an error
	err2 := rotatingWriter.Close()
	if err2 != nil {
		t.Errorf("Closing already closed writer returned error: %v", err2)
	}

	// Verify the file still contains the data
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read log file after close: %v", err)
	} else if !strings.Contains(string(content), "Before close") {
		t.Error("File content lost after close")
	}
}

// TestRotatorWithLargeData tests writing large amounts of data
func TestRotatorWithLargeData(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_large.log")

	rotationConfig := &config.RotationConfig{
		MaxSize: 500, // Small size to trigger rotation with large data
	}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	defer rotatingWriter.Close()

	// Write larger data to potentially trigger rotation logic (though the current implementation doesn't actually rotate)
	largeData := make([]byte, 400) // Close to max size
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	n, err := rotatingWriter.Write(largeData)
	if err != nil {
		t.Errorf("Write large data returned error: %v", err)
	}
	if n != len(largeData) {
		t.Errorf("Write large data returned %d, expected %d", n, len(largeData))
	}

	// Write more data
	moreData := make([]byte, 200)
	for i := range moreData {
		moreData[i] = byte('B' + (i % 26))
	}

	// This write might cause rotation in a full implementation
	n2, err2 := rotatingWriter.Write(moreData)
	if err2 != nil {
		t.Errorf("Second write returned error: %v", err2)
	}
	if n2 != len(moreData) {
		t.Errorf("Second write returned %d, expected %d", n2, len(moreData))
	}

	// Check final file size
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	} else if int64(len(content)) < int64(len(largeData)+len(moreData)) {
		t.Errorf("File size %d is smaller than expected total %d", len(content), len(largeData)+len(moreData))
	}
}

// TestRotatorWithSpecialCharacters tests writing with special characters
func TestRotatorWithSpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_special.log")

	rotationConfig := &config.RotationConfig{}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	defer rotatingWriter.Close()

	// Write data with special characters
	specialData := []byte("Log with special chars: \n \t \r \" ' & < > 日本語 Ελληνικά\n")

	n, err := rotatingWriter.Write(specialData)
	if err != nil {
		t.Errorf("Write special characters returned error: %v", err)
	}
	if n != len(specialData) {
		t.Errorf("Write special characters returned %d, expected %d", n, len(specialData))
	}

	// Check that the file contains the special data
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
	} else if string(content) != string(specialData) {
		t.Errorf("File content doesn't match special data exactly")
	}
}

// TestRotatorFilePermissions tests file permissions
func TestRotatorFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_perms.log")

	rotationConfig := &config.RotationConfig{}

	rotatingWriter, err := NewRotator(tempFile, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotator returned error: %v", err)
	}
	defer rotatingWriter.Close()

	// Write some data
	_, _ = rotatingWriter.Write([]byte("test content\n"))

	// Check file permissions
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Errorf("Failed to stat log file: %v", err)
	} else {
		// On Unix systems, the default mode for os.OpenFile with O_CREATE is 0666
		// but the actual permissions depend on umask, so we just check that it's a regular file
		if !info.Mode().IsRegular() {
			t.Error("Log file is not a regular file")
		}
	}
}

// at
func TestRotatorWithNonExistentDir(t *testing.T) {
	// at
	rotationConfig := &config.RotationConfig{}

	nonExistentFile := "/non/existent/dir/test.log"
	rotatingWriter, err := NewRotator(nonExistentFile, rotationConfig)

	if err == nil {
		t.Error("NewRotator should have failed for non-existent directory")
		if rotatingWriter != nil {
			rotatingWriter.Close()
		}
		return
	}

	if rotatingWriter != nil {
		t.Error("NewRotator should have returned nil for non-existent directory")
		rotatingWriter.Close()
	}
}
