package util

import (
	"bytes"
	"sync"
	"testing"
)

// TestGetBufferFromPool tests the GetBufferFromPool function
func TestGetBufferFromPool(t *testing.T) {
	// Get a buffer from the pool
	buf1 := GetBufferFromPool()
	if buf1 == nil {
		t.Fatal("GetBufferFromPool returned nil")
	}
	
	// Verify initial state
	if buf1.Len() != 0 {
		t.Errorf("New buffer from pool should have length 0, got %d", buf1.Len())
	}
	
	// Write some data to the buffer
	_, err := buf1.WriteString("test data")
	if err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}
	
	// Return the buffer to the pool
	PutBufferToPool(buf1)
	
	// Get another buffer - it might be the same one we just returned
	buf2 := GetBufferFromPool()
	if buf2 == nil {
		t.Fatal("GetBufferFromPool returned nil after return")
	}
	
	// The buffer should be reset to empty state
	if buf2.Len() != 0 {
		t.Errorf("Returned buffer from pool should have length 0, got %d", buf2.Len())
	}
	
	// Return the buffer to the pool again
	PutBufferToPool(buf2)
}

// TestPutBufferToPool tests the PutBufferToPool function
func TestPutBufferToPool(t *testing.T) {
	buf := GetBufferFromPool()
	
	// Write data to the buffer
	buf.WriteString("some data")
	
	// Verify it has content
	if buf.Len() == 0 {
		t.Fatal("Buffer should have content before putting to pool")
	}
	
	// Put it back to pool - this should reset it
	PutBufferToPool(buf)
	
	// Get it again to verify it's reset
	buf2 := GetBufferFromPool()
	defer PutBufferToPool(buf2)
	
	if buf2.Len() != 0 {
		t.Errorf("Returned buffer should be empty, got length %d", buf2.Len())
	}
}

// TestGetSmallByteSliceFromPool tests the GetSmallByteSliceFromPool function
func TestGetSmallByteSliceFromPool(t *testing.T) {
	slice1 := GetSmallByteSliceFromPool()
	if slice1 == nil {
		t.Fatal("GetSmallByteSliceFromPool returned nil")
	}
	
	// Initially, the slice should have 0 length but some capacity
	if len(slice1) != 0 {
		t.Errorf("New slice from pool should have length 0, got %d", len(slice1))
	}
	if cap(slice1) == 0 {
		t.Error("New slice from pool should have some capacity")
	}
	
	// Add some data
	slice1 = append(slice1, []byte("test")...)
	if len(slice1) != 4 {
		t.Errorf("Slice should have length 4 after appending 'test', got %d", len(slice1))
	}
	
	// Return to pool
	PutSmallByteSliceToPool(slice1)
	
	// Get another slice
	slice2 := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(slice2)
	
	// It should be reset to 0 length
	if len(slice2) != 0 {
		t.Errorf("Returned slice from pool should have length 0, got %d", len(slice2))
	}
}

// TestPutSmallByteSliceToPool tests the PutSmallByteSliceToPool function
func TestPutSmallByteSliceToPool(t *testing.T) {
	slice := GetSmallByteSliceFromPool()
	
	// Add some data
	slice = append(slice, []byte("data")...)
	
	// Put it back
	PutSmallByteSliceToPool(slice)
	
	// Get it again to check if it's properly reset
	slice2 := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(slice2)
	
	if len(slice2) != 0 {
		t.Errorf("Returned slice should be empty, got length %d", len(slice2))
	}
}

// TestGetMapStringFromPool tests the GetMapStringFromPool function
func TestGetMapStringFromPool(t *testing.T) {
	map1 := GetMapStringFromPool()
	if map1 == nil {
		t.Fatal("GetMapStringFromPool returned nil")
	}
	
	// Initially, the map should be empty
	if len(map1) != 0 {
		t.Errorf("New map from pool should have length 0, got %d", len(map1))
	}
	
	// Add some data
	map1["key1"] = "value1"
	map1["key2"] = "value2"
	if len(map1) != 2 {
		t.Errorf("Map should have length 2 after adding items, got %d", len(map1))
	}
	
	// Return to pool
	PutMapStringToPool(map1)
	
	// Get another map
	map2 := GetMapStringFromPool()
	defer PutMapStringToPool(map2)
	
	// It should be reset to empty
	if len(map2) != 0 {
		t.Errorf("Returned map from pool should have length 0, got %d", len(map2))
	}
}

// TestPutMapStringToPool tests the PutMapStringToPool function
func TestPutMapStringToPool(t *testing.T) {
	m := GetMapStringFromPool()
	
	// Add some data
	m["test"] = "value"
	
	// Put it back to pool
	PutMapStringToPool(m)
	
	// Get it again to check if it's properly reset
	m2 := GetMapStringFromPool()
	defer PutMapStringToPool(m2)
	
	if len(m2) != 0 {
		t.Errorf("Returned map should be empty, got length %d", len(m2))
	}
}

// TestGetStringSliceFromPool tests the GetStringSliceFromPool function
func TestGetStringSliceFromPool(t *testing.T) {
	slice1 := GetStringSliceFromPool()
	if slice1 == nil {
		t.Fatal("GetStringSliceFromPool returned nil")
	}

	// Initially, the slice should have 0 length but some capacity
	if len(slice1) != 0 {
		t.Errorf("New slice from pool should have length 0, got %d", len(slice1))
	}

	// Add some data
	slice1 = append(slice1, "item1", "item2")
	if len(slice1) != 2 {
		t.Errorf("Slice should have length 2 after appending items, got %d", len(slice1))
	}

	// Return to pool
	PutStringSliceToPool(slice1)

	// Get another slice
	slice2 := GetStringSliceFromPool()
	defer PutStringSliceToPool(slice2)

	// It should be reset to 0 length
	if len(slice2) != 0 {
		t.Errorf("Returned slice from pool should have length 0, got %d", len(slice2))
	}
}

// TestPutStringSliceToPool tests the PutStringSliceToPool function
func TestPutStringSliceToPool(t *testing.T) {
	slice := GetStringSliceFromPool()

	// Add some data
	slice = append(slice, "test")

	// Put it back
	PutStringSliceToPool(slice)

	// Get it again to check if it's properly reset
	slice2 := GetStringSliceFromPool()
	defer PutStringSliceToPool(slice2)

	if len(slice2) != 0 {
		t.Errorf("Returned slice should be empty, got length %d", len(slice2))
	}
}

// TestPoolMetrics tests the PoolMetrics functionality
func TestPoolMetrics(t *testing.T) {
	metrics := GetPoolMetrics()
	
	initialBufferGetCount := metrics.BufferGetCount()
	initialBufferPutCount := metrics.BufferPutCount()
	initialSliceGetCount := metrics.SliceGetCount()
	initialSlicePutCount := metrics.SlicePutCount()
	initialMapGetCount := metrics.MapGetCount()
	initialMapPutCount := metrics.MapPutCount()
	
	// Perform some pool operations
	buf := GetBufferFromPool()
	PutBufferToPool(buf)
	
	slice := GetSmallByteSliceFromPool()
	PutSmallByteSliceToPool(slice)
	
	m := GetMapStringFromPool()
	PutMapStringToPool(m)
	
	s := GetStringSliceFromPool()
	PutStringSliceToPool(s)
	
	newBufferGetCount := metrics.BufferGetCount()
	newBufferPutCount := metrics.BufferPutCount()
	newSliceGetCount := metrics.SliceGetCount()
	newSlicePutCount := metrics.SlicePutCount()
	newMapGetCount := metrics.MapGetCount()
	newMapPutCount := metrics.MapPutCount()
	
	if newBufferGetCount != initialBufferGetCount+1 {
		t.Error("BufferGetCount was not incremented properly")
	}
	if newBufferPutCount != initialBufferPutCount+1 {
		t.Error("BufferPutCount was not incremented properly")
	}
	if newSliceGetCount != initialSliceGetCount+2 {
		t.Errorf("SliceGetCount was not incremented properly: expected %d, got %d", initialSliceGetCount+2, newSliceGetCount)
	}
	if newSlicePutCount != initialSlicePutCount+2 {
		t.Errorf("SlicePutCount was not incremented properly: expected %d, got %d", initialSlicePutCount+2, newSlicePutCount)
	}
	if newMapGetCount != initialMapGetCount+1 {
		t.Error("MapGetCount was not incremented properly")
	}
	if newMapPutCount != initialMapPutCount+1 {
		t.Error("MapPutCount was not incremented properly")
	}
}

// TestPoolMetricsConcurrent tests the PoolMetrics functionality in a concurrent context
func TestPoolMetricsConcurrent(t *testing.T) {
	metrics := GetPoolMetrics()
	
	initialGetCount := metrics.BufferGetCount()
	
	// Run multiple goroutines to use the pool concurrently
	const numGoroutines = 10
	const operationsPerGoroutine = 100
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				buf := GetBufferFromPool()
				PutBufferToPool(buf)
			}
		}()
	}
	
	wg.Wait()
	
	finalGetCount := metrics.BufferGetCount()
	
	expected := initialGetCount + int64(numGoroutines*operationsPerGoroutine)
	if finalGetCount != expected {
		t.Errorf("Concurrent BufferGetCount: expected %d, got %d", expected, finalGetCount)
	}
}

// TestLogBuffer tests the LogBuffer type
func TestLogBuffer(t *testing.T) {
	buf := &LogBuffer{
		buf: make([]byte, 0, 100),
		len: 0,
	}
	
	// Test WriteBytes
	err := buf.WriteBytes([]byte("hello"))
	if err != nil {
		t.Errorf("WriteBytes returned error: %v", err)
	}
	
	if buf.len != 5 {
		t.Errorf("WriteBytes should update length to 5, got %d", buf.len)
	}
	
	if string(buf.buf[:buf.len]) != "hello" {
		t.Errorf("Buffer content should be 'hello', got '%s'", string(buf.buf[:buf.len]))
	}
	
	// Test WriteByte
	err = buf.WriteByte(' ')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	err = buf.WriteByte('w')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	err = buf.WriteByte('o')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	err = buf.WriteByte('r')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	err = buf.WriteByte('l')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	err = buf.WriteByte('d')
	if err != nil {
		t.Errorf("WriteByte returned error: %v", err)
	}
	
	expected := "hello world"
	if string(buf.buf[:buf.len]) != expected {
		t.Errorf("Buffer content should be '%s', got '%s'", expected, string(buf.buf[:buf.len]))
	}
	
	// Test Bytes
	result := buf.Bytes()
	if string(result) != expected {
		t.Errorf("Bytes() should return '%s', got '%s'", expected, string(result))
	}
	
	// Test Reset
	buf.Reset()
	if buf.len != 0 {
		t.Errorf("Reset should set length to 0, got %d", buf.len)
	}
	
	// Test available
	if buf.available() != cap(buf.buf) {
		t.Errorf("available should return capacity %d, got %d", cap(buf.buf), buf.available())
	}
}

// TestLogBufferFull tests the LogBuffer when it's full
func TestLogBufferFull(t *testing.T) {
	// Create a buffer with limited capacity
	buf := &LogBuffer{
		buf: make([]byte, 5), // capacity 5
		len: 5,               // already at capacity
	}
	
	// Try to write - should return error
	err := buf.WriteBytes([]byte("extra data"))
	if err == nil {
		t.Error("WriteBytes should return error when buffer is full")
	}
	
	// Try to write a single byte - should return error
	err = buf.WriteByte('x')
	if err == nil {
		t.Error("WriteByte should return error when buffer is full")
	}
}

// TestSmallByteSlicePool tests the behavior of small byte slice pool with size limits
func TestSmallByteSlicePool(t *testing.T) {
	// Get a slice and make it larger than MaxSmallSlicePoolSize
	// This should cause it to not be returned to the pool in PutSmallByteSliceToPool
	largeSlice := make([]byte, MaxSmallSlicePoolSize+10) // Larger than the limit
	
	// Put it to pool - this should result in the slice being discarded
	PutSmallByteSliceToPool(largeSlice)
	
	// Check if the discard counter increased
	// We'd need to check internal pool state which is difficult
	// For now, just ensure it doesn't panic
}

// TestGoroutineLocalBufferPool tests the goroutine-local buffer pool
func TestGoroutineLocalBufferPool(t *testing.T) {
	// Get the local pool for the current goroutine
	localPool := GetGoroutineLocalBufferPool()
	if localPool == nil {
		t.Fatal("GetGoroutineLocalBufferPool returned nil")
	}
	
	// Test getting from local pool
	_ = localPool.GetBufferFromLocalPool()
	// buf might be nil if the local pool is empty, which is expected

	// Put a buffer to local pool
	testBuf := bytes.NewBuffer(make([]byte, 0, 100))
	_ = localPool.PutBufferToLocalPool(testBuf)
	// returned might be false if the local pool is full, which is expected
}

// TestPutBufferToLocalPoolFull tests what happens when the local pool is full
func TestPutBufferToLocalPoolFull(t *testing.T) {
	localPool := GetGoroutineLocalBufferPool()
	
	// Fill up the local pool's channel
	for i := 0; i < 10; i++ { // Default channel size is 10
		buf := bytes.NewBuffer(make([]byte, 0, 100))
		returned := localPool.PutBufferToLocalPool(buf)
		// If returned is false, it means the local pool was full and it was put to global pool
		if !returned {
			// This is acceptable behavior
			break // Exit the loop if the local pool is full
		}
		// If we're able to put all 10, that's also fine
	}
	
	// Try to put one more - this should return false and put to global pool
	extraBuf := bytes.NewBuffer(make([]byte, 0, 100))
	_ = localPool.PutBufferToLocalPool(extraBuf)
	// This might return false if local pool is full, which is expected behavior
}