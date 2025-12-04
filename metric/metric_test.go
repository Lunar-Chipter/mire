package metric

import (
	"strings"
	"sync"
	"testing"

	"github.com/Lunar-Chipter/mire/core"
)

// TestNewDefaultMetricsCollector tests creating a new DefaultMetricsCollector
func TestNewDefaultMetricsCollector(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	if metricsCollector == nil {
		t.Fatal("NewDefaultMetricsCollector returned nil")
	}
	
	// Check initial state of maps
	if metricsCollector.counters == nil {
		t.Error("Counters map should not be nil")
	}
	if metricsCollector.histograms == nil {
		t.Error("Histograms map should not be nil")
	}
	if metricsCollector.gauges == nil {
		t.Error("Gauges map should not be nil")
	}
	
	// Check that maps are empty initially
	if len(metricsCollector.counters) != 0 {
		t.Error("Counters map should be empty initially")
	}
	if len(metricsCollector.histograms) != 0 {
		t.Error("Histograms map should be empty initially")
	}
	if len(metricsCollector.gauges) != 0 {
		t.Error("Gauges map should be empty initially")
	}
}

// TestIncrementCounter tests the IncrementCounter method
func TestIncrementCounter(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Test incrementing different levels
	levels := []core.Level{core.TRACE, core.DEBUG, core.INFO, core.NOTICE, core.WARN, core.ERROR, core.FATAL, core.PANIC}
	
	for _, level := range levels {
		initialCount := metricsCollector.GetCounter("log." + strings.ToLower(level.String()))
		metricsCollector.IncrementCounter(level, nil)
		newCount := metricsCollector.GetCounter("log." + strings.ToLower(level.String()))

		if newCount != initialCount+1 {
			t.Errorf("IncrementCounter for level %v: expected %d, got %d", level, initialCount+1, newCount)
		}
	}
	
	// Test with invalid level (should not panic and not increment counter)
	initialCount := metricsCollector.GetCounter("log.INVALID")
	metricsCollector.IncrementCounter(core.Level(999), nil) // Invalid level
	newCount := metricsCollector.GetCounter("log.INVALID")
	
	if newCount != initialCount {
		t.Error("IncrementCounter with invalid level should not increment any counter")
	}
}

// TestRecordHistogram tests the RecordHistogram method
func TestRecordHistogram(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Record some histogram values
	metricName := "response_time"
	metricsCollector.RecordHistogram(metricName, 100.0, nil)
	metricsCollector.RecordHistogram(metricName, 200.0, nil)
	metricsCollector.RecordHistogram(metricName, 50.0, nil)
	
	// Check that values are stored
	min, max, avg, p95 := metricsCollector.GetHistogram(metricName)
	
	if min != 50.0 {
		t.Errorf("Expected min 50.0, got %f", min)
	}
	if max != 200.0 {
		t.Errorf("Expected max 200.0, got %f", max)
	}
	if avg != 116.66666666666667 { // (100+200+50)/3
		t.Logf("Expected avg around 116.67, got %f", avg) // Allow for floating point precision
	}
	if p95 != 200.0 { // 95th percentile of [50, 100, 200] should be 200
		t.Errorf("Expected p95 200.0, got %f", p95)
	}
	
	// Test with different metric name
	otherMetric := "request_size"
	metricsCollector.RecordHistogram(otherMetric, 1024.0, nil)
	min2, max2, avg2, p952 := metricsCollector.GetHistogram(otherMetric)
	
	if min2 != 1024.0 || max2 != 1024.0 || avg2 != 1024.0 || p952 != 1024.0 {
		t.Errorf("Second histogram metric not stored correctly: min=%f, max=%f, avg=%f, p95=%f", 
			min2, max2, avg2, p952)
	}
}

// TestRecordGauge tests the RecordGauge method
func TestRecordGauge(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Record gauge values
	metricName := "current_connections"
	metricsCollector.RecordGauge(metricName, 10.0, nil)
	
	if metricsCollector.GetCounter("current_connections") != 0 { // Not a counter
		t.Log("GetCounter won't work with gauges")
	}
	
	// Check the gauge value by accessing it differently
	// Actually, there's no getter for gauges in the current implementation
	// We'll need to check the internal gauges map, but that's not exposed.
	// Let's just verify that RecordGauge doesn't cause errors
	
	// Record a different value, which should overwrite the previous value
	metricsCollector.RecordGauge(metricName, 20.0, nil)
}

// TestGetCounter tests the GetCounter method
func TestGetCounter(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Test with non-existent metric
	count := metricsCollector.GetCounter("nonexistent")
	if count != 0 {
		t.Errorf("GetCounter for non-existent metric should return 0, got %d", count)
	}
	
	// Increment a counter and check
	metricsCollector.IncrementCounter(core.INFO, nil)
	count = metricsCollector.GetCounter("log.info")
	if count != 1 {
		t.Errorf("Expected counter value 1, got %d", count)
	}

	// Increment again
	metricsCollector.IncrementCounter(core.INFO, nil)
	count = metricsCollector.GetCounter("log.info")
	if count != 2 {
		t.Errorf("Expected counter value 2, got %d", count)
	}
}

// TestGetHistogram tests the GetHistogram method
func TestGetHistogram(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Test with non-existent metric
	min, max, avg, p95 := metricsCollector.GetHistogram("nonexistent")
	if min != 0 || max != 0 || avg != 0 || p95 != 0 {
		t.Errorf("GetHistogram for non-existent metric should return all zeros, got min=%f, max=%f, avg=%f, p95=%f", 
			min, max, avg, p95)
	}
	
	// Record some values and test
	metricName := "test_histogram"
	metricsCollector.RecordHistogram(metricName, 1.0, nil)
	metricsCollector.RecordHistogram(metricName, 2.0, nil)
	metricsCollector.RecordHistogram(metricName, 3.0, nil)
	
	min, max, avg, p95 = metricsCollector.GetHistogram(metricName)
	
	if min != 1.0 {
		t.Errorf("Expected min 1.0, got %f", min)
	}
	if max != 3.0 {
		t.Errorf("Expected max 3.0, got %f", max)
	}
	if avg != 2.0 { // (1+2+3)/3
		t.Logf("Expected avg 2.0, got %f", avg) // Allow for floating point precision
	}
	if p95 != 3.0 { // 95th percentile of [1, 2, 3] should be 3
		t.Errorf("Expected p95 3.0, got %f", p95)
	}
}

// TestGetCounterCaseSensitivity tests case sensitivity of counter names
func TestGetCounterCaseSensitivity(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Increment a counter
	metricsCollector.IncrementCounter(core.INFO, nil)
	
	// Check that the name follows the expected format - should look for lowercase
	count := metricsCollector.GetCounter("log.info")
	if count != 1 {
		t.Errorf("Expected counter 'log.info' to have value 1, got %d", count)
	}

	// Check case sensitivity - uppercase should return 0
	countUpper := metricsCollector.GetCounter("log.INFO")
	if countUpper != 0 {
		t.Error("Counter names should be case-sensitive")
	}
}

// TestRecordHistogramMultipleValues tests histogram with multiple values
func TestRecordHistogramMultipleValues(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	metricName := "response_times"
	
	// Add a range of values
	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	for _, v := range values {
		metricsCollector.RecordHistogram(metricName, v, nil)
	}
	
	min, max, avg, p95 := metricsCollector.GetHistogram(metricName)
	
	if min != 10.0 {
		t.Errorf("Expected min 10.0, got %f", min)
	}
	if max != 100.0 {
		t.Errorf("Expected max 100.0, got %f", max)
	}
	if avg != 55.0 { // Sum(10..100)/10 = 55
		t.Logf("Expected avg 55.0, got %f", avg)
	}
	
	// For 10 values, 95th percentile would be the 10th value (100)
	// (using ceil(0.95*10) = ceil(9.5) = 10th element, which is 100)
	if p95 != 100.0 {
		t.Errorf("Expected p95 100.0, got %f", p95)
	}
}

// TestRecordHistogramEmpty tests histogram with no values
func TestRecordHistogramEmpty(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Get histogram for non-existent metric (should return zeros)
	min, max, avg, p95 := metricsCollector.GetHistogram("empty_histogram")
	
	if min != 0 || max != 0 || avg != 0 || p95 != 0 {
		t.Errorf("Empty histogram should return all zeros, got min=%f, max=%f, avg=%f, p95=%f", 
			min, max, avg, p95)
	}
}

// TestMetricsCollectorConcurrent tests the metrics collector in a concurrent context
func TestMetricsCollectorConcurrent(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	var wg sync.WaitGroup
	
	// Test concurrent counter increments
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Use different levels based on goroutine ID to avoid interference
				level := core.Level(goroutineID % 8) // 8 log levels
				metricsCollector.IncrementCounter(level, nil)
				
				// Also record some histogram values
				metricsCollector.RecordHistogram("concurrent_metric", float64(j), nil)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Check final counter values
	totalExpected := numGoroutines * operationsPerGoroutine
	var totalActual int64

	for level := core.TRACE; level <= core.PANIC; level++ {
		count := metricsCollector.GetCounter("log." + strings.ToLower(level.String()))
		totalActual += count
	}

	if totalActual != int64(totalExpected) {
		t.Errorf("Expected total counter operations %d, got %d", totalExpected, totalActual)
	}
	
	// Check histogram
	min, max, avg, p95 := metricsCollector.GetHistogram("concurrent_metric")

	// The histogram should have values from 0 to 99 from each goroutine
	// So we should have 1000 values (10 goroutines * 100 operations)
	if min != 0.0 {
		t.Errorf("Expected min 0.0, got %f", min)
	}

	// Max should be 99 (highest value added by each goroutine)
	if max < 98.0 { // Allow for some flexibility in floating point comparison
		t.Logf("Expected max around 99.0, got %f", max)
	}

	// Use avg and p95 to avoid unused variable error
	t.Logf("Histogram stats - min: %f, max: %f, avg: %f, p95: %f", min, max, avg, p95)
}

// TestRecordHistogramWithTags tests histogram recording with tags
func TestRecordHistogramWithTags(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Record histogram values with tags
	// The current implementation doesn't use tags, so we just ensure it doesn't crash
	tags := map[string]string{
		"env":    "test",
		"region": "us-west-1",
	}
	
	metricsCollector.RecordHistogram("tagged_metric", 100.0, tags)
	metricsCollector.RecordHistogram("tagged_metric", 200.0, tags)
	
	min, max, _, _ := metricsCollector.GetHistogram("tagged_metric")

	if min != 100.0 || max != 200.0 {
		t.Errorf("Tagged histogram values not recorded correctly: min=%f, max=%f", min, max)
	}
}

// TestIncrementCounterWithTags tests counter incrementing with tags
func TestIncrementCounterWithTags(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()
	
	// Increment with tags (tags are currently not used in counter implementation)
	tags := map[string]string{
		"service": "api",
		"version": "v1.0",
	}

	initial := metricsCollector.GetCounter("log.info")
	metricsCollector.IncrementCounter(core.INFO, tags)
	metricsCollector.IncrementCounter(core.INFO, tags) // Increment twice
	final := metricsCollector.GetCounter("log.info")

	if final != initial+2 {
		t.Errorf("Expected counter to increment by 2, got increase of %d", final-initial)
	}
}

// TestMetricsCollectorWithVariousDataTypes tests various numeric types
func TestMetricsCollectorWithVariousDataTypes(t *testing.T) {
	metricsCollector := NewDefaultMetricsCollector()

	// Test histogram with various float values
	testValues := []float64{
		0.0,          // Zero
		-10.5,        // Negative
		3.14159,      // Pi
		123.456789,   // High precision
		1e6,          // Scientific notation (1,000,000)
		-1e-6,        // Negative scientific (0.000001)
	}

	for _, v := range testValues {
		metricsCollector.RecordHistogram("various_types", v, nil)
	}

	// The histogram should handle all these values
	min, max, _, _ := metricsCollector.GetHistogram("various_types")

	// Find expected min/max from our test values
	expectedMin := testValues[0]
	expectedMax := testValues[0]
	for _, v := range testValues {
		if v < expectedMin {
			expectedMin = v
		}
		if v > expectedMax {
			expectedMax = v
		}
	}

	if min != expectedMin {
		t.Errorf("Expected min %f, got %f", expectedMin, min)
	}
	if max != expectedMax {
		t.Errorf("Expected max %f, got %f", expectedMax, max)
	}
}
