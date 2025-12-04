package util

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Goroutine pool for managing clock workers
var clockWorkerPool = sync.Pool{
	New: func() interface{} {
		return make(chan time.Time, 16) // Pre-allocated buffered channel
	},
}

// Buffer pool for zero-allocation time formatting
var timeBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 32) // Pre-allocated buffer
	},
}

// Metrics for observability
type ClockMetrics struct {
	updateCount  atomic.Int64
	errorCount   atomic.Int64
	lastUpdate   atomic.Int64
}

// Clock provides a clock
// by updating an atomic time variable in a background goroutine.
type Clock struct {
	atomicTime atomic.Value // Stores time.Time
	interval   time.Duration
	stop       chan struct{}
	wg         sync.WaitGroup
	metrics    *ClockMetrics
	_          [64 - unsafe.Sizeof(time.Duration(0))]byte // Padding for cache alignment
}

// Constants for compile-time configuration
const (
	DefaultInterval = time.Millisecond * 10 // Standard update interval
	FastInterval    = time.Millisecond      // Frequent update interval
	SlowInterval    = time.Second           // Infrequent update interval
)

// NewClock creates and starts a new clock.
func NewClock(interval time.Duration) *Clock {
	fc := &Clock{
		interval: interval,
		stop:     make(chan struct{}, 1), // Buffered channel to prevent blocking
		metrics:  &ClockMetrics{},
	}
	fc.atomicTime.Store(time.Now()) // Initialize with current time

	if interval > 0 {
		fc.wg.Add(1)
		go fc.run()
	}
	return fc
}

// run updates the atomic time at the specified interval.
func (c *Clock) run() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Get a worker from the pool
	worker := clockWorkerPool.Get().(chan time.Time)
	defer clockWorkerPool.Put(worker)

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.atomicTime.Store(now)
			c.metrics.updateCount.Add(1)
			c.metrics.lastUpdate.Store(now.UnixNano())

			// Send to worker pool for processing
			select {
			case worker <- now:
			default:
				// Worker busy, increment error count
				c.metrics.errorCount.Add(1)
			}
		case <-c.stop:
			return
		default:
			// Prevent tight loop, yield control
			time.Sleep(time.Microsecond)
		}
	}
}

// Now returns the current time from the clock.
func (c *Clock) Now() time.Time {
	return c.atomicTime.Load().(time.Time)
}

// Stop stops the clock's background goroutine.
func (c *Clock) Stop() {
	if c.interval > 0 {
		// Use a select statement to avoid panic if channel is already closed
		select {
		case <-c.stop:
			// Channel is already closed, nothing to do
		default:
			close(c.stop)
		}
		c.wg.Wait()
	}
}

// Metrics returns the clock's metrics
func (c *Clock) Metrics() *ClockMetrics {
	return c.metrics
}

// UpdateCount returns the number of times the clock has been updated
func (cm *ClockMetrics) UpdateCount() int64 {
	return cm.updateCount.Load()
}

// LastUpdate returns the timestamp of the last update
func (cm *ClockMetrics) LastUpdate() int64 {
	return cm.lastUpdate.Load()
}

// ErrorCount returns the number of errors encountered
func (cm *ClockMetrics) ErrorCount() int64 {
	return cm.errorCount.Load()
}

// Manual byte manipulation for time formatting
func (c *Clock) TimeToBytes() []byte {
	t := c.Now()
	// Get pre-allocated buffer from pool for zero allocation
	buf := timeBufferPool.Get().([]byte)
	buf = buf[:0] // Reset buffer length without reallocating

	// Manual formatting to avoid fmt overhead
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	// Format: YYYY-MM-DD HH:MM:SS
	buf = append(buf,
		byte(year/1000)+'0', byte((year/100)%10)+'0', byte((year/10)%10)+'0', byte(year%10)+'0',
		'-',
		byte(month/10)+'0', byte(month%10)+'0',
		'-',
		byte(day/10)+'0', byte(day%10)+'0',
		' ',
		byte(hour/10)+'0', byte(hour%10)+'0',
		':',
		byte(min/10)+'0', byte(min%10)+'0',
		':',
		byte(sec/10)+'0', byte(sec%10)+'0',
	)

	return buf
}

// Global clock instance
var globalClock = NewClock(DefaultInterval)

// Now returns the current time from the global clock
func Now() time.Time {
	return globalClock.Now()
}

// ReleaseTimeBuffer returns the buffer to the pool after use
func (c *Clock) ReleaseTimeBuffer(buf []byte) {
	timeBufferPool.Put(buf[:0]) // Reset before putting back
}
