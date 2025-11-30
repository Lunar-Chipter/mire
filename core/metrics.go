package core

import (
	"sync/atomic"
	"time"
)

// CoreMetrics menyimpan metrik untuk observability
// Built-in performance metrics dengan zero allocation overhead
type CoreMetrics struct {
	// Entry-related metrics
	EntryCreatedCount    atomic.Int64
	EntryReusedCount     atomic.Int64
	EntryPoolMissCount   atomic.Int64
	EntrySerializedCount atomic.Int64
	
	// Buffer-related metrics
	BufferGetCount       atomic.Int64
	BufferPutCount       atomic.Int64
	BufferMissCount      atomic.Int64
	
	// Slice-related metrics
	SliceGetCount        atomic.Int64
	SlicePutCount        atomic.Int64
	
	// Error metrics
	ErrorCount           atomic.Int64
	
	// Timing metrics
	LastOperationTime    atomic.Int64
	ProcessingTime       atomic.Int64
}

// Global metrics instance
var globalCoreMetrics = &CoreMetrics{}

// GetCoreMetrics mengembalikan instance global metrics
func GetCoreMetrics() *CoreMetrics {
	return globalCoreMetrics
}

// Entry metrics methods
func (cm *CoreMetrics) IncEntryCreated() {
	cm.EntryCreatedCount.Add(1)
}

func (cm *CoreMetrics) IncEntryReused() {
	cm.EntryReusedCount.Add(1)
}

func (cm *CoreMetrics) IncEntryPoolMiss() {
	cm.EntryPoolMissCount.Add(1)
}

func (cm *CoreMetrics) IncEntrySerialized() {
	cm.EntrySerializedCount.Add(1)
}

// Buffer metrics methods
func (cm *CoreMetrics) IncBufferGet() {
	cm.BufferGetCount.Add(1)
}

func (cm *CoreMetrics) IncBufferPut() {
	cm.BufferPutCount.Add(1)
}

func (cm *CoreMetrics) IncBufferMiss() {
	cm.BufferMissCount.Add(1)
}

// Slice metrics methods
func (cm *CoreMetrics) IncSliceGet() {
	cm.SliceGetCount.Add(1)
}

func (cm *CoreMetrics) IncSlicePut() {
	cm.SlicePutCount.Add(1)
}

// Error metrics methods
func (cm *CoreMetrics) IncError() {
	cm.ErrorCount.Add(1)
}

// Timing methods
func (cm *CoreMetrics) SetLastOperationTime(t time.Time) {
	cm.LastOperationTime.Store(t.UnixNano())
}

func (cm *CoreMetrics) AddProcessingTime(duration time.Duration) {
	cm.ProcessingTime.Add(int64(duration))
}

// GetEntryMetrics mengembalikan statistik entry
func (cm *CoreMetrics) GetEntryMetrics() map[string]int64 {
	return map[string]int64{
		"created":    cm.EntryCreatedCount.Load(),
		"reused":     cm.EntryReusedCount.Load(),
		"pool_miss":  cm.EntryPoolMissCount.Load(),
		"serialized": cm.EntrySerializedCount.Load(),
		"hit_ratio":  cm.EntryReusedCount.Load() * 100 / (cm.EntryCreatedCount.Load() + cm.EntryReusedCount.Load() + 1), // +1 untuk menghindari pembagian dengan 0
	}
}

// GetBufferMetrics mengembalikan statistik buffer
func (cm *CoreMetrics) GetBufferMetrics() map[string]int64 {
	return map[string]int64{
		"gets":      cm.BufferGetCount.Load(),
		"puts":      cm.BufferPutCount.Load(),
		"misses":    cm.BufferMissCount.Load(),
		"hit_ratio": (cm.BufferGetCount.Load() - cm.BufferMissCount.Load()) * 100 / (cm.BufferGetCount.Load() + 1), // +1 untuk menghindari pembagian dengan 0
	}
}

// GetSliceMetrics mengembalikan statistik slice
func (cm *CoreMetrics) GetSliceMetrics() map[string]int64 {
	return map[string]int64{
		"gets": cm.SliceGetCount.Load(),
		"puts": cm.SlicePutCount.Load(),
	}
}

// GetErrorMetrics mengembalikan statistik error
func (cm *CoreMetrics) GetErrorMetrics() map[string]int64 {
	return map[string]int64{
		"errors": cm.ErrorCount.Load(),
	}
}

// GetTimingMetrics mengembalikan statistik timing
func (cm *CoreMetrics) GetTimingMetrics() map[string]int64 {
	return map[string]int64{
		"processing_time_ns": cm.ProcessingTime.Load(),
		"last_operation":     cm.LastOperationTime.Load(),
	}
}

// GetAllMetrics mengembalikan semua metrik
func (cm *CoreMetrics) GetAllMetrics() map[string]interface{} {
	allMetrics := make(map[string]interface{})
	
	allMetrics["entries"] = cm.GetEntryMetrics()
	allMetrics["buffers"] = cm.GetBufferMetrics()
	allMetrics["slices"] = cm.GetSliceMetrics()
	allMetrics["errors"] = cm.GetErrorMetrics()
	allMetrics["timing"] = cm.GetTimingMetrics()
	
	return allMetrics
}

// ResetMetrics mereset semua metrik ke 0
func (cm *CoreMetrics) ResetMetrics() {
	cm.EntryCreatedCount.Store(0)
	cm.EntryReusedCount.Store(0)
	cm.EntryPoolMissCount.Store(0)
	cm.EntrySerializedCount.Store(0)
	
	cm.BufferGetCount.Store(0)
	cm.BufferPutCount.Store(0)
	cm.BufferMissCount.Store(0)
	
	cm.SliceGetCount.Store(0)
	cm.SlicePutCount.Store(0)
	
	cm.ErrorCount.Store(0)
	
	cm.LastOperationTime.Store(0)
	cm.ProcessingTime.Store(0)
}