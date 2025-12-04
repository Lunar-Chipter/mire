package config

import (
	"testing"
	"time"
)

// TestRotationConfig tests the RotationConfig struct
func TestRotationConfig(t *testing.T) {
	// Create a RotationConfig with various values
	rotationConfig := &RotationConfig{
		MaxSize:         1024 * 1024, // 1MB
		MaxAge:          7 * 24 * time.Hour, // 7 days
		MaxBackups:      10,
		LocalTime:       true,
		Compress:        false,
		RotationTime:    24 * time.Hour, // 1 day
		FilenamePattern: "2006-01-02", // Go time format
	}
	
	// Verify all values are set correctly
	if rotationConfig.MaxSize != 1024*1024 {
		t.Errorf("Expected MaxSize 1MB, got %d", rotationConfig.MaxSize)
	}
	
	if rotationConfig.MaxAge != 7*24*time.Hour {
		t.Errorf("Expected MaxAge 7 days, got %v", rotationConfig.MaxAge)
	}
	
	if rotationConfig.MaxBackups != 10 {
		t.Errorf("Expected MaxBackups 10, got %d", rotationConfig.MaxBackups)
	}
	
	if rotationConfig.LocalTime != true {
		t.Errorf("Expected LocalTime true, got %v", rotationConfig.LocalTime)
	}
	
	if rotationConfig.Compress != false {
		t.Errorf("Expected Compress false, got %v", rotationConfig.Compress)
	}
	
	if rotationConfig.RotationTime != 24*time.Hour {
		t.Errorf("Expected RotationTime 24 hours, got %v", rotationConfig.RotationTime)
	}
	
	if rotationConfig.FilenamePattern != "2006-01-02" {
		t.Errorf("Expected FilenamePattern '2006-01-02', got %s", rotationConfig.FilenamePattern)
	}
}

// TestRotationConfigZeroValues tests RotationConfig with zero values
func TestRotationConfigZeroValues(t *testing.T) {
	// Create a RotationConfig with zero values
	rotationConfig := &RotationConfig{}
	
	// Verify zero values are handled correctly
	if rotationConfig.MaxSize != 0 {
		t.Errorf("Expected zero MaxSize, got %d", rotationConfig.MaxSize)
	}
	
	if rotationConfig.MaxAge != 0 {
		t.Errorf("Expected zero MaxAge, got %v", rotationConfig.MaxAge)
	}
	
	if rotationConfig.MaxBackups != 0 {
		t.Errorf("Expected zero MaxBackups, got %d", rotationConfig.MaxBackups)
	}
	
	if rotationConfig.LocalTime != false {
		t.Errorf("Expected LocalTime false, got %v", rotationConfig.LocalTime)
	}
	
	if rotationConfig.Compress != false {
		t.Errorf("Expected Compress false, got %v", rotationConfig.Compress)
	}
	
	if rotationConfig.RotationTime != 0 {
		t.Errorf("Expected zero RotationTime, got %v", rotationConfig.RotationTime)
	}
	
	if rotationConfig.FilenamePattern != "" {
		t.Errorf("Expected empty FilenamePattern, got %s", rotationConfig.FilenamePattern)
	}
}

// TestRotationConfigWithDifferentSizes tests different size configurations
func TestRotationConfigWithDifferentSizes(t *testing.T) {
	testCases := []struct {
		name     string
		maxSize  int64
		expected int64
	}{
		{"1KB", 1024, 1024},
		{"1MB", 1024 * 1024, 1024 * 1024},
		{"10MB", 10 * 1024 * 1024, 10 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024, 100 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024, 1024 * 1024 * 1024},
		{"Zero", 0, 0},
		{"Negative", -1, -1},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				MaxSize: tc.maxSize,
			}
			
			if config.MaxSize != tc.expected {
				t.Errorf("Expected MaxSize %d, got %d", tc.expected, config.MaxSize)
			}
		})
	}
}

// TestRotationConfigWithDifferentDurations tests different duration configurations
func TestRotationConfigWithDifferentDurations(t *testing.T) {
	testCases := []struct {
		name   string
		maxAge time.Duration
	}{
		{"1 hour", time.Hour},
		{"6 hours", 6 * time.Hour},
		{"12 hours", 12 * time.Hour},
		{"1 day", 24 * time.Hour},
		{"3 days", 3 * 24 * time.Hour},
		{"1 week", 7 * 24 * time.Hour},
		{"Zero", 0},
		{"Negative", -1 * time.Hour},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				MaxAge: tc.maxAge,
			}
			
			if config.MaxAge != tc.maxAge {
				t.Errorf("Expected MaxAge %v, got %v", tc.maxAge, config.MaxAge)
			}
		})
	}
}

// TestRotationConfigWithDifferentBackups tests different backup counts
func TestRotationConfigWithDifferentBackups(t *testing.T) {
	testCases := []struct {
		name     string
		maxBackups int
	}{
		{"Zero", 0},
		{"One", 1},
		{"Five", 5},
		{"Ten", 10},
		{"Hundred", 100},
		{"Negative", -1},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				MaxBackups: tc.maxBackups,
			}
			
			if config.MaxBackups != tc.maxBackups {
				t.Errorf("Expected MaxBackups %d, got %d", tc.maxBackups, config.MaxBackups)
			}
		})
	}
}

// TestRotationConfigBooleans tests boolean configuration options
func TestRotationConfigBooleans(t *testing.T) {
	testCases := []struct {
		name      string
		localTime bool
		compress  bool
	}{
		{"Both false", false, false},
		{"LocalTime only", true, false},
		{"Compress only", false, true},
		{"Both true", true, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				LocalTime: tc.localTime,
				Compress:  tc.compress,
			}
			
			if config.LocalTime != tc.localTime {
				t.Errorf("Expected LocalTime %v, got %v", tc.localTime, config.LocalTime)
			}
			
			if config.Compress != tc.compress {
				t.Errorf("Expected Compress %v, got %v", tc.compress, config.Compress)
			}
		})
	}
}

// TestRotationConfigWithTimeDurations tests different time durations
func TestRotationConfigWithTimeDurations(t *testing.T) {
	// Test RotationTime field with different values
	testCases := []struct {
		name         string
		rotationTime time.Duration
	}{
		{"1 minute", time.Minute},
		{"5 minutes", 5 * time.Minute},
		{"15 minutes", 15 * time.Minute},
		{"30 minutes", 30 * time.Minute},
		{"1 hour", time.Hour},
		{"2 hours", 2 * time.Hour},
		{"6 hours", 6 * time.Hour},
		{"12 hours", 12 * time.Hour},
		{"24 hours", 24 * time.Hour},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				RotationTime: tc.rotationTime,
			}
			
			if config.RotationTime != tc.rotationTime {
				t.Errorf("Expected RotationTime %v, got %v", tc.rotationTime, config.RotationTime)
			}
		})
	}
}

// TestRotationConfigWithFilenamePatterns tests different filename patterns
func TestRotationConfigWithFilenamePatterns(t *testing.T) {
	testCases := []struct {
		name            string
		filenamePattern string
	}{
		{"Default Go time format", "2006-01-02"},
		{"Extended format", "2006-01-02T15-04-05"},
		{"Year-Month", "2006-01"},
		{"Day of year", "2006-002"},
		{"Unix timestamp", "Unix"},
		{"Custom pattern", "log-2006-01-02"},
		{"Empty pattern", ""},
		{"Complex pattern", "app-2006-01-02-15-04-05.log"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &RotationConfig{
				FilenamePattern: tc.filenamePattern,
			}
			
			if config.FilenamePattern != tc.filenamePattern {
				t.Errorf("Expected FilenamePattern %s, got %s", tc.filenamePattern, config.FilenamePattern)
			}
		})
	}
}

// TestRotationConfigLargeValues tests with very large values
func TestRotationConfigLargeValues(t *testing.T) {
	// Test with very large values that could be used in production
	config := &RotationConfig{
		MaxSize:    100 * 1024 * 1024 * 1024, // 100GB
		MaxAge:     365 * 24 * time.Hour,      // 1 year
		MaxBackups: 1000,                      // 1000 backups
	}
	
	if config.MaxSize != 100*1024*1024*1024 {
		t.Errorf("Expected very large MaxSize, got %d", config.MaxSize)
	}
	
	if config.MaxAge != 365*24*time.Hour {
		t.Errorf("Expected very long MaxAge, got %v", config.MaxAge)
	}
	
	if config.MaxBackups != 1000 {
		t.Errorf("Expected large MaxBackups, got %d", config.MaxBackups)
	}
}

// TestRotationConfigMemoryLayout tests the memory layout of RotationConfig
func TestRotationConfigMemoryLayout(t *testing.T) {
	// This test checks that the struct has reasonable size and layout
	config := &RotationConfig{}
	
	// We can't easily test the exact memory layout without unsafe,
	// but we can ensure the struct can be instantiated and used
	if config == nil {
		t.Error("RotationConfig could not be instantiated")
	}
	
	// Check that the struct has the expected fields by setting and getting them
	config.MaxSize = 12345
	config.MaxAge = 456 * time.Hour
	config.MaxBackups = 789
	config.LocalTime = true
	config.Compress = true
	config.RotationTime = 10 * time.Hour
	config.FilenamePattern = "test"
	
	// Verify the values are still there
	if config.MaxSize != 12345 {
		t.Error("MaxSize was not set correctly")
	}
	if config.MaxAge != 456*time.Hour {
		t.Error("MaxAge was not set correctly")
	}
	if config.MaxBackups != 789 {
		t.Error("MaxBackups was not set correctly")
	}
	if config.LocalTime != true {
		t.Error("LocalTime was not set correctly")
	}
	if config.Compress != true {
		t.Error("Compress was not set correctly")
	}
	if config.RotationTime != 10*time.Hour {
		t.Error("RotationTime was not set correctly")
	}
	if config.FilenamePattern != "test" {
		t.Error("FilenamePattern was not set correctly")
	}
}

// TestRotationConfigComparison tests comparing config values
func TestRotationConfigComparison(t *testing.T) {
	config1 := &RotationConfig{
		MaxSize:         1024,
		MaxAge:          time.Hour,
		MaxBackups:      5,
		LocalTime:       true,
		Compress:        false,
		RotationTime:    time.Minute,
		FilenamePattern: "test",
	}
	
	config2 := &RotationConfig{
		MaxSize:         1024,
		MaxAge:          time.Hour,
		MaxBackups:      5,
		LocalTime:       true,
		Compress:        false,
		RotationTime:    time.Minute,
		FilenamePattern: "test",
	}
	
	// Since RotationConfig contains time.Duration fields, direct comparison with == won't work
	// We'll compare fields individually
	if config1.MaxSize != config2.MaxSize {
		t.Error("Configs should be equal but MaxSize differs")
	}
	if config1.MaxAge != config2.MaxAge {
		t.Error("Configs should be equal but MaxAge differs")
	}
	if config1.MaxBackups != config2.MaxBackups {
		t.Error("Configs should be equal but MaxBackups differs")
	}
	if config1.LocalTime != config2.LocalTime {
		t.Error("Configs should be equal but LocalTime differs")
	}
	if config1.Compress != config2.Compress {
		t.Error("Configs should be equal but Compress differs")
	}
	if config1.RotationTime != config2.RotationTime {
		t.Error("Configs should be equal but RotationTime differs")
	}
	if config1.FilenamePattern != config2.FilenamePattern {
		t.Error("Configs should be equal but FilenamePattern differs")
	}
}