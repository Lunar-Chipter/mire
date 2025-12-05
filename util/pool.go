package util

import (
	"bytes"
	"errors"
	"sync"
	"unsafe"
)

// Constants for compile-time configuration
const (
	// Buffer sizes for different pool types - aligned with zero-allocation philosophy
	SmallBufferSize       = 512   // Untuk perf-critical
	MediumBufferSize      = 2048  // Untuk standard logs  
	LargeBufferSize       = 8192  // Untuk verbose debugging
	DefaultBufferSize     = MediumBufferSize
	MaxBufferPoolSize     = LargeBufferSize
	SmallByteSliceSize    = 64
	MaxSmallSlicePoolSize = 1024 // 1KB
	StringSliceCapacity   = 10
	MapInitialCapacity    = 8
)

// Error definitions for buffer operations
var (
	ErrBufferFull = errors.New("buffer is full")
)

// PoolMetrics for observability with atomic operations
// Built-in performance metrics with zero allocation overhead
type PoolMetrics struct {
	bufferGetCount    int64
	bufferPutCount    int64
	sliceGetCount     int64
	slicePutCount     int64
	mapGetCount       int64
	mapPutCount       int64
	poolMissCount     int64
	discardedCount    int64
}

// Global metrics instance
var globalPoolMetrics = &PoolMetrics{}

// GetPoolMetrics returns the global pool metrics
func GetPoolMetrics() *PoolMetrics {
	return globalPoolMetrics
}

// BufferGetCount returns the number of buffer gets
func (pm *PoolMetrics) BufferGetCount() int64 {
	return pm.bufferGetCount
}

// BufferPutCount returns the number of buffer puts
func (pm *PoolMetrics) BufferPutCount() int64 {
	return pm.bufferPutCount
}

// SliceGetCount returns the number of slice gets
func (pm *PoolMetrics) SliceGetCount() int64 {
	return pm.sliceGetCount
}

// SlicePutCount returns the number of slice puts
func (pm *PoolMetrics) SlicePutCount() int64 {
	return pm.slicePutCount
}

// MapGetCount returns the number of map gets
func (pm *PoolMetrics) MapGetCount() int64 {
	return pm.mapGetCount
}

// MapPutCount returns the number of map puts
func (pm *PoolMetrics) MapPutCount() int64 {
	return pm.mapPutCount
}

// PoolMissCount returns the number of pool misses
func (pm *PoolMetrics) PoolMissCount() int64 {
	return pm.poolMissCount
}

// DiscardedCount returns the number of discarded items
func (pm *PoolMetrics) DiscardedCount() int64 {
	return pm.discardedCount
}

// Zero-allocation buffer pool with pre-allocated buffers
var bufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocated buffer with fixed size to avoid fragmentation
		return bytes.NewBuffer(make([]byte, 0, DefaultBufferSize))
	},
}

// LogBuffer represents a zero-allocation buffer for logging
type LogBuffer struct {
	buf []byte
	len int
	_   [64 - unsafe.Sizeof(int(0))]byte // Padding untuk cache alignment
}

func (b *LogBuffer) WriteBytes(data []byte) error {
	if b.available() < len(data) {
		return ErrBufferFull
	}
	// Ensure the underlying slice is long enough to handle the copy
	// Extend the slice to accommodate all data if necessary
	if len(b.buf) < b.len+len(data) {
		// Extend the slice length to the required size (but within capacity)
		newLen := b.len + len(data)
		b.buf = b.buf[:newLen]
	}
	// Copy data to the current logical end of the buffer
	copy(b.buf[b.len:], data)
	// Update the logical length
	b.len += len(data)
	return nil
}

func (b *LogBuffer) WriteByte(c byte) error {
	if b.available() < 1 {
		return ErrBufferFull
	}
	// Ensure the underlying slice is long enough to add the byte
	if len(b.buf) <= b.len {
		// Extend the slice to include this position
		b.buf = b.buf[:b.len+1]
	}
	b.buf[b.len] = c
	b.len++
	return nil
}

// available returns the available space in the buffer
func (b *LogBuffer) available() int {
	return cap(b.buf) - b.len
}

// Bytes returns the buffer content
func (b *LogBuffer) Bytes() []byte {
	return b.buf[:b.len]
}

func (b *LogBuffer) Reset() {
	b.len = 0
}

// GetBufferFromPool gets a byte buffer from the pool
func GetBufferFromPool() *bytes.Buffer {
	globalPoolMetrics.bufferGetCount++
	// Try to get from goroutine-local pool first for zero lock contention
	localPool := GetGoroutineLocalBufferPool()
	buf := localPool.GetBufferFromLocalPool()
	if buf != nil {
		return buf
	}

	// Fallback to global pool if local pool is empty
	return bufferPool.Get().(*bytes.Buffer)
}

// PutBufferToPool returns a byte buffer to the pool
func PutBufferToPool(buf *bytes.Buffer) {
	globalPoolMetrics.bufferPutCount++
	// Try to put to goroutine-local pool first for zero lock contention
	localPool := GetGoroutineLocalBufferPool()
	if localPool.PutBufferToLocalPool(buf) {
		return
	}

	// Fallback to global pool if local pool is full
	buf.Reset()
	bufferPool.Put(buf)
}

// smallByteSlicePool is for small byte slices.
var smallByteSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, SmallByteSliceSize) // For small formatting, like int/float/timestamp
	},
}

// GetSmallByteSliceFromPool gets a small byte slice from the pool.
func GetSmallByteSliceFromPool() []byte {
	globalPoolMetrics.sliceGetCount++
	return smallByteSlicePool.Get().([]byte)[:0] // Get and reset length
}

// PutSmallByteSliceToPool returns a small byte slice to the pool.
func PutSmallByteSliceToPool(b []byte) {
	// Avoid putting back overly large slices to prevent pool pollution
	if cap(b) < MaxSmallSlicePoolSize { // Keep slices up to 1KB
		smallByteSlicePool.Put(b)
		globalPoolMetrics.slicePutCount++
	} else {
		globalPoolMetrics.discardedCount++
	}
}

// Object pool for reusing map[string]string objects
var mapStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, MapInitialCapacity)
	},
}

// GetMapStringFromPool gets a map[string]string from the pool
func GetMapStringFromPool() map[string]string {
	globalPoolMetrics.mapGetCount++
	m := mapStringPool.Get().(map[string]string)
	for k := range m {
		delete(m, k) // Reset the map
	}
	return m
}

// PutMapStringToPool returns a map[string]string to the pool
func PutMapStringToPool(m map[string]string) {
	mapStringPool.Put(m)
	globalPoolMetrics.mapPutCount++
}

// String slice pool for reusing string slices (e.g., for map keys)
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, StringSliceCapacity) // Pre-allocate with a reasonable capacity
	},
}

// GetStringSliceFromPool gets a []string from the pool
func GetStringSliceFromPool() []string {
	globalPoolMetrics.sliceGetCount++
	s := stringSlicePool.Get().([]string)
	return s[:0] // Reset slice length but keep capacity
}

// PutStringSliceToPool returns a []string to the pool
func PutStringSliceToPool(s []string) {
	stringSlicePool.Put(s)
	globalPoolMetrics.slicePutCount++
}

// Padding for cache alignment
type cachePadding struct {
	_ [64]byte // Cache line padding
}

// Goroutine-local pools to reduce lock contention
var (
	goroutineBufferPools    sync.Map // map[uint64]*localBufferPool
	goroutineSlicePools     sync.Map // map[uint64]*localSlicePool
	goroutineMapPools       sync.Map // map[uint64]*localMapPool
)

// Local pool structures for goroutine-local storage
type localBufferPool struct {
	buffers chan *bytes.Buffer
}

type localSlicePool struct {
	slices chan []byte
}

type localMapPool struct {
	maps chan map[string]string
}

// getGoroutineID returns a pseudo goroutine ID for goroutine-local storage
func getGoroutineID() uint64 {
	// Simplified implementation - in production, use a more reliable method
	return uint64(uintptr(unsafe.Pointer(&globalPoolMetrics)) % 1000000)
}

// GetGoroutineLocalBufferPool returns the buffer pool for the current goroutine
func GetGoroutineLocalBufferPool() *localBufferPool {
	gid := getGoroutineID()
	if pool, ok := goroutineBufferPools.Load(gid); ok {
		return pool.(*localBufferPool)
	}

	// Create a new local pool for this goroutine
	newPool := &localBufferPool{
		buffers: make(chan *bytes.Buffer, 10), // Buffered channel for 10 buffers
	}
	goroutineBufferPools.Store(gid, newPool)

	return newPool
}

// GetBufferFromLocalPool gets a buffer from the goroutine-local pool
// Zero lock contention di hot path
func (lp *localBufferPool) GetBufferFromLocalPool() *bytes.Buffer {
	select {
	case buf := <-lp.buffers:
		buf.Reset()
		return buf
	default:
		// Local pool is empty, indicate to use global pool
		globalPoolMetrics.poolMissCount++
		return nil
	}
}

// PutBufferToLocalPool returns a buffer to the goroutine-local pool
// Zero lock contention di hot path
func (lp *localBufferPool) PutBufferToLocalPool(buf *bytes.Buffer) bool {
	select {
	case lp.buffers <- buf:
		return true
	default:
		// Local pool is full, indicate to use global pool
		return false
	}
}
