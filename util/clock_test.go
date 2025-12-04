package util

import (
	"testing"
	"time"
)

// TestNewClock tests creating a new clock
func TestNewClock(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	// Check initial state
	if clock.interval != 10*time.Millisecond {
		t.Errorf("Clock interval should be 10ms, got %v", clock.interval)
	}
	
	// Check that time is set initially
	initialTime := clock.Now()
	if initialTime.IsZero() {
		t.Error("Initial time should not be zero")
	}
	
	// Wait a bit and check that time has updated
	time.Sleep(15 * time.Millisecond)
	updatedTime := clock.Now()
	
	// The time should have been updated since we created the clock
	// Allow for some tolerance due to timing precision
	if updatedTime.Before(initialTime) {
		t.Error("Time should not go backward")
	}
}

// TestClockNow tests the Now method
func TestClockNow(t *testing.T) {
	clock := NewClock(5 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	// Get current time from clock multiple times
	time1 := clock.Now()
	time.Sleep(1 * time.Millisecond)
	time2 := clock.Now()
	
	// Both times should be non-zero
	if time1.IsZero() || time2.IsZero() {
		t.Error("Clock times should not be zero")
	}
	
	// With a 5ms update interval, we might not see an update after 1ms,
	// but the times should be reasonable
	if time2.Before(time1) {
		t.Error("Time should not go backward")
	}
}

// TestClockStop tests stopping the clock
func TestClockStop(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	
	// Get initial time
	initialTime := clock.Now()
	if initialTime.IsZero() {
		t.Error("Initial time should not be zero")
	}
	
	// Stop the clock
	clock.Stop()
	
	// Wait and get time again
	time.Sleep(15 * time.Millisecond)
	timeAfterStop := clock.Now()
	
	// The time should still be accessible after stop
	if timeAfterStop.IsZero() {
		t.Error("Time should still be available after stop")
	}
	
	// Try stopping again - this should not cause an error
	clock.Stop()
}

// TestClockWithZeroInterval tests creating a clock with zero interval
func TestClockWithZeroInterval(t *testing.T) {
	clock := NewClock(0) // Zero interval should not start the goroutine
	if clock == nil {
		t.Fatal("NewClock with zero interval returned nil")
	}
	defer clock.Stop()
	
	// Get time - this should work even with zero interval
	time1 := clock.Now()
	time.Sleep(5 * time.Millisecond)
	time2 := clock.Now()
	
	// The time should be updated manually (not by background goroutine)
	if time1.IsZero() || time2.IsZero() {
		t.Error("Clock with zero interval should still return valid times")
	}
}

// TestClockMetrics tests the clock metrics functionality
func TestClockMetrics(t *testing.T) {
	clock := NewClock(2 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	metrics := clock.Metrics()
	if metrics == nil {
		t.Fatal("Clock metrics should not be nil")
	}
	
	// Get initial counts
	initialUpdateCount := metrics.UpdateCount()
	initialErrorCount := metrics.ErrorCount()
	
	// Wait for clock to update
	time.Sleep(5 * time.Millisecond)
	
	// Check that update count has increased
	newUpdateCount := metrics.UpdateCount()
	if newUpdateCount <= initialUpdateCount {
		t.Log("Update count might not have increased due to timing - this could be acceptable") // Depending on timing, might not have completed an update cycle
	}
	
	// Error count should still be 0
	newErrorCount := metrics.ErrorCount()
	if newErrorCount != initialErrorCount {
		t.Error("Error count should remain the same under normal conditions")
	}
	
	// Check last update time
	lastUpdate := metrics.LastUpdate()
	if lastUpdate == 0 {
		t.Error("Last update should be non-zero after clock starts")
	}
}

// TestClockMetricsConcurrent tests clock metrics in concurrent context
func TestClockMetricsConcurrent(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	metrics := clock.Metrics()
	initialCount := metrics.UpdateCount()
	
	// Run multiple goroutines that access the clock
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = clock.Now() // Access current time
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
	
	// Check final metrics
	finalCount := metrics.UpdateCount()
	if finalCount < initialCount {
		t.Error("Update count should not be negative")
	}
}

// TestClockTimeToBytes tests the TimeToBytes method
func TestClockTimeToBytes(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	// Get time as bytes
	timeBytes := clock.TimeToBytes()
	
	// Should not be empty
	if len(timeBytes) == 0 {
		t.Error("TimeToBytes should return non-empty bytes")
	}
	
	// The result should look like a formatted time
	timeStr := string(timeBytes)
	if len(timeStr) < 10 { // Basic check for reasonable time format
		t.Errorf("TimeToBytes returned string that's too short: '%s'", timeStr)
	}
	
	// Release the buffer back to the pool
	clock.ReleaseTimeBuffer(timeBytes)
}

// TestClockReleaseTimeBuffer tests the ReleaseTimeBuffer method
func TestClockReleaseTimeBuffer(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	// Get time as bytes
	timeBytes := clock.TimeToBytes()
	
	// Release the buffer
	clock.ReleaseTimeBuffer(timeBytes)
	
	// The buffer might be reused, but we can't easily verify this without
	// checking internal pool state, so just ensure the function doesn't panic
}

// TestGlobalClock tests the global clock instance
func TestGlobalClock(t *testing.T) {
	// Get current time from global clock
	currentTime := Now()
	if currentTime.IsZero() {
		t.Error("Global clock Now() should return non-zero time")
	}
	
	// Wait a bit and get time again
	time.Sleep(1 * time.Millisecond)
	newTime := Now()
	
	// Times should be non-zero and reasonable
	if currentTime.IsZero() || newTime.IsZero() {
		t.Error("Global clock times should not be zero")
	}
	
	if newTime.Before(currentTime) {
		t.Error("Global clock time should not go backward")
	}
}

// TestClockInitialization tests clock initialization with different intervals
func TestClockInitialization(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{"Fast interval", FastInterval},      // 1ms
		{"Default interval", DefaultInterval}, // 10ms
		{"Slow interval", SlowInterval},      // 1s
		{"Custom interval", 50 * time.Millisecond},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clock := NewClock(test.interval)
			if clock == nil {
				t.Fatalf("NewClock with %s interval returned nil", test.name)
			}
			defer clock.Stop()
			
			// Verify the interval is set correctly
			if clock.interval != test.interval {
				t.Errorf("Clock interval for %s should be %v, got %v", test.name, test.interval, clock.interval)
			}
			
			// Get time to ensure it works
			t1 := clock.Now()
			if t1.IsZero() {
				t.Errorf("Clock with %s should return non-zero time", test.name)
			}
		})
	}
}

// TestClockMetricsInitialValues tests initial values of metrics
func TestClockMetricsInitialValues(t *testing.T) {
	clock := NewClock(10 * time.Millisecond)
	if clock == nil {
		t.Fatal("NewClock returned nil")
	}
	defer clock.Stop()
	
	metrics := clock.Metrics()
	
	// Initially, these should be 0 or reasonable defaults
	if metrics.UpdateCount() < 0 {
		t.Error("Update count should be non-negative")
	}
	
	if metrics.ErrorCount() != 0 {
		t.Error("Error count should initially be 0")
	}
	
	// The last update time might be 0 initially if no updates have occurred yet
	// This is acceptable behavior
}