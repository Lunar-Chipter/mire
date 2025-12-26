package core

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// LogEntry represents a single log entry with all its metadata
type LogEntry struct {
	Timestamp        time.Time                                `json:"timestamp"`                // When the log was created
	Level            Level                                    `json:"level"`                    // Log severity level
	LevelName        []byte                                   `json:"level_name"`               // Byte representation of level for zero-allocation formatting
	Message          []byte                                   `json:"message"`                  // Log message
	Caller           *Caller                                  `json:"caller,omitempty"`         // Caller information
	Fields           map[string][]byte                        `json:"fields,omitempty"`         // Additional fields as []byte for zero allocation
	PID              int                                      `json:"pid"`                      // Process ID
	GoroutineID      []byte                                   `json:"goroutine_id,omitempty"`   // Goroutine ID as byte slice
	TraceID          []byte                                   `json:"trace_id,omitempty"`       // Trace ID for distributed tracing as byte slice
	SpanID           []byte                                   `json:"span_id,omitempty"`        // Span ID for distributed tracing as byte slice
	UserID           []byte                                   `json:"user_id,omitempty"`        // User ID as byte slice
	SessionID        []byte                                   `json:"session_id,omitempty"`     // Session ID as byte slice
	RequestID        []byte                                   `json:"request_id,omitempty"`     // Request ID as byte slice
	Duration         time.Duration                            `json:"duration,omitempty"`       // Operation duration
	Error            error                                    `json:"error,omitempty"`          // Error information
	StackTrace       []byte                                   `json:"stack_trace,omitempty"`    // Stack trace
	StackTraceBufPtr *[]byte                                  `json:"-"`                        // Pointer to the pooled buffer for StackTrace
	Hostname         []byte                                   `json:"hostname,omitempty"`       // Hostname as byte slice
	Application      []byte                                   `json:"application,omitempty"`    // Application name as byte slice
	Version          []byte                                   `json:"version,omitempty"`        // Application version as byte slice
	Environment      []byte                                   `json:"environment,omitempty"`    // Environment (dev/prod/etc) as byte slice
	CustomMetrics    map[string]float64                       `json:"custom_metrics,omitempty"` // Custom metrics
	Tags             [][]byte                                 `json:"tags,omitempty"`           // Tags for categorization as byte slices
	KeyVals          [][]byte                                 `json:"keyvals,omitempty"`        // Zero-allocation key-value pairs
	_                [64 - unsafe.Sizeof(time.Time{})%64]byte // Padding for cache alignment
}

// Reset resets the LogEntry for reuse
func (e *LogEntry) Reset() {
	e.Level = 0
	e.Message = nil
	e.Timestamp = time.Time{}
	e.Caller = nil
	e.Fields = nil
	e.KeyVals = nil
	e.TraceID = nil
	e.SpanID = nil
	e.UserID = nil
	e.RequestID = nil
	e.SessionID = nil
	e.Hostname = nil
	e.Application = nil
	e.Version = nil
	e.Environment = nil
	e.CustomMetrics = nil
	e.Tags = nil
	e.Duration = 0
	e.Error = nil
	e.StackTrace = nil
	e.GoroutineID = nil
	e.PID = 0
	e.LevelName = nil
	e.StackTraceBufPtr = nil
}

// Caller contains information about the code location where the log was created
type Caller struct {
	File     string `json:"file"`     // Source file name
	Line     int    `json:"line"`     // Line number
	Function string `json:"function"` // Function name
	Package  string `json:"package"`  // Package name
}

// Object pool for reusing Caller objects
var callerInfoPool = sync.Pool{
	New: func() interface{} {
		return &Caller{}
	},
}

// GetCallerFromPool gets a Caller from the pool
func GetCallerFromPool() *Caller {
	return callerInfoPool.Get().(*Caller)
}

// PutCallerToPool returns a Caller to the pool
func PutCallerToPool(ci *Caller) {
	// Reset fields to avoid data leakage
	ci.File = ""
	ci.Line = 0
	ci.Function = ""
	ci.Package = ""
	callerInfoPool.Put(ci)
}

// Object pool for reusing map[string]float64 objects
var mapFloatPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]float64)
	},
}

// GetMapFloatFromPool gets a map[string]float64 from the pool
func GetMapFloatFromPool() map[string]float64 {
	m := mapFloatPool.Get().(map[string]float64)
	for k := range m {
		delete(m, k) // Reset the map
	}
	return m
}

// PutMapFloatToPool returns a map[string]float64 to the pool
func PutMapFloatToPool(m map[string]float64) {
	mapFloatPool.Put(m)
}

// Object pool for reusing map[string][]byte objects
var mapBytePool = sync.Pool{
	New: func() interface{} {
		return make(map[string][]byte)
	},
}

// GetMapByteFromPool gets a map[string][]byte from the pool
func GetMapByteFromPool() map[string][]byte {
	m := mapBytePool.Get().(map[string][]byte)
	for k := range m {
		delete(m, k) // Reset the map
	}
	return m
}

// PutMapByteToPool returns a map[string][]byte to the pool
func PutMapByteToPool(m map[string][]byte) {
	mapBytePool.Put(m)
}

// stringSlicePool is a pool for reusing string slices
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, TagsSliceCapacity)
		return &s
	},
}

// bufferPool is a pool for reusing byte buffers for serialization
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, MediumEntryBufferSize)
		return &buf
	},
}

// GetBuffer gets a byte buffer from the pool
func GetBuffer() *[]byte {
	buf := bufferPool.Get().(*[]byte)
	*buf = (*buf)[:0] // Reset length but keep capacity
	return buf
}

// PutBuffer returns a byte buffer to the pool
func PutBuffer(buf *[]byte) {
	bufferPool.Put(buf)
}

// GetStringSliceFromPool gets a string slice from the pool
func GetStringSliceFromPool() *[]string {
	s := stringSlicePool.Get().(*[]string)
	*s = (*s)[:0] // to
	return s
}

// PutStringSliceToPool returns a string slice to the pool
func PutStringSliceToPool(s *[]string) {
	stringSlicePool.Put(s)
}

// Global metrics instance - now using CoreMetrics
var globalEntryMetrics = GetCoreMetrics()

// GetEntryMetrics returns the global entry metrics
func GetEntryMetrics() *CoreMetrics {
	return globalEntryMetrics
}

// CreatedCount returns the number of entries created
func (em *CoreMetrics) CreatedCount() int64 {
	return em.EntryCreatedCount.Load()
}

// ReusedCount returns the number of entries reused from pool
func (em *CoreMetrics) ReusedCount() int64 {
	return em.EntryReusedCount.Load()
}

// PoolMissCount returns the number of pool misses
func (em *CoreMetrics) PoolMissCount() int64 {
	return em.EntryPoolMissCount.Load()
}

// SerializedCount returns the number of entries serialized
func (em *CoreMetrics) SerializedCount() int64 {
	return em.EntrySerializedCount.Load()
}

// clearMap clears a map
func clearMap(m map[string][]byte) {
	for k := range m {
		delete(m, k)
	}
}

// clearFloatMap clears a map of floats
func clearFloatMap(m map[string]float64) {
	for k := range m {
		delete(m, k)
	}
}

// clearByteSliceSlice clears a byte slice slice
func clearByteSliceSlice(s [][]byte) [][]byte {
	return s[:0]
}

// Object pool for reusing LogEntry objects
var entryPool = sync.Pool{
	New: func() interface{} {
		tags := GetStringSliceFromPool()
		tagsAsBytes := make([][]byte, 0, len(*tags))
		for _, tag := range *tags {
			tagsAsBytes = append(tagsAsBytes, StringToBytes(tag))
		}
		return &LogEntry{
			Fields:        GetMapByteFromPool(),
			CustomMetrics: GetMapFloatFromPool(),
			Tags:          tagsAsBytes,
		}
	},
}

// Constants for compile-time configuration
const (
	// Buffer sizes for different log entry types
	SmallEntryBufferSize  = 256
	MediumEntryBufferSize = 1024
	LargeEntryBufferSize  = 4096

	// Pre-allocated slice capacities
	TagsSliceCapacity  = 10
	FieldsMapCapacity  = 8
	MetricsMapCapacity = 4

	// Performance optimization constants
	PreallocatedPoolSize     = 32 // Number of pre-allocated entries
	MaxOptimizedPathAttempts = 5  // Max attempts to use optimized path before fallback
)

// GetEntryFromPool gets a LogEntry from the pool
func GetEntryFromPool() *LogEntry {
	// Fallback: Use regular goroutine-local pool
	localPool := GetGoroutineLocalEntryPool()
	return localPool.GetLocalEntry()
}

// GetGlobalEntry gets a LogEntry directly from the global pool
func GetGlobalEntry() *LogEntry {
	entry := entryPool.Get().(*LogEntry)

	// Update metrics
	if entry.Timestamp.IsZero() {
		// New entry created (pool miss)
		globalEntryMetrics.IncEntryPoolMiss()
		globalEntryMetrics.IncEntryCreated()
	} else {
		// Entry reused
		globalEntryMetrics.IncEntryReused()
	}

	// Reset fields to avoid data leakage
	entry.Timestamp = time.Time{}
	entry.Level = INFO
	entry.LevelName = nil
	entry.Message = nil
	entry.Caller = nil
	clearMap(entry.Fields)
	clearFloatMap(entry.CustomMetrics)
	entry.Tags = clearByteSliceSlice(entry.Tags)
	entry.PID = 0
	entry.GoroutineID = nil
	entry.TraceID = nil
	entry.SpanID = nil
	entry.UserID = nil
	entry.SessionID = nil
	entry.RequestID = nil
	entry.Duration = 0
	entry.Error = nil
	entry.StackTrace = nil
	entry.Hostname = nil
	entry.Application = nil
	entry.Version = nil
	entry.Environment = nil

	return entry
}

// goroutineLocalEntryPool stores entry pool per goroutine to avoid lock contention
var goroutineLocalEntryPool = sync.Map{}

// poolLastAccess tracks last access time for each goroutine pool
var poolLastAccess = sync.Map{}

// Cleanup configuration
const (
	poolMaxAgeThreshold = 10 * time.Minute
	poolCleanupInterval = 5 * time.Minute
	maxGoroutinePools   = 1000
)

// StartPoolCleanup starts periodic cleanup of goroutine-local pools
func StartPoolCleanup() {
	go func() {
		ticker := time.NewTicker(poolCleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			cleanupOldPools()
		}
	}()
}

// cleanupOldPools removes pools that haven't been accessed recently
func cleanupOldPools() {
	now := time.Now()
	threshold := now.Add(-poolMaxAgeThreshold)

	poolLastAccess.Range(func(key, value interface{}) bool {
		lastAccess := value.(time.Time)
		if lastAccess.Before(threshold) {
			goroutineLocalEntryPool.Delete(key)
			poolLastAccess.Delete(key)
		}
		return true
	})

	// Limit total number of pools to prevent unbounded growth
	var poolCount int
	goroutineLocalEntryPool.Range(func(key, value interface{}) bool {
		poolCount++
		if poolCount > maxGoroutinePools {
			goroutineLocalEntryPool.Delete(key)
			poolLastAccess.Delete(key)
		}
		return true
	})
}

// LocalPool represents a per-goroutine entry pool
type LocalPool struct {
	entries chan *LogEntry
}

// getGoroutineID gets the current goroutine ID by parsing runtime.Stack
func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	if n == 0 {
		return 0
	}

	// Parse "goroutine 123 [running]:" format
	str := string(buf[:n])

	// Find "goroutine "
	idx := strings.Index(str, "goroutine ")
	if idx == -1 {
		return 0
	}

	// Find space after ID
	idStart := idx + 10 // length of "goroutine "
	endIdx := strings.Index(str[idStart:], " ")
	if endIdx == -1 {
		return 0
	}

	// Parse ID
	idStr := str[idStart : idStart+endIdx]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

// GetGoroutineLocalEntryPool gets entry pool for current goroutine
func GetGoroutineLocalEntryPool() *LocalPool {
	gid := getGoroutineID()
	now := time.Now()

	if pool, ok := goroutineLocalEntryPool.Load(gid); ok {
		// Update last access time
		poolLastAccess.Store(gid, now)
		return pool.(*LocalPool)
	}

	// Create new pool for this goroutine
	newPool := &LocalPool{
		entries: make(chan *LogEntry, 100),
	}
	goroutineLocalEntryPool.Store(gid, newPool)
	poolLastAccess.Store(gid, now)

	return newPool
}

// GetLocalEntry gets entry from local goroutine pool
func (g *LocalPool) GetLocalEntry() *LogEntry {
	select {
	case entry := <-g.entries:
		// Reset fields to avoid data leakage
		entry.Timestamp = time.Time{}
		entry.Level = INFO
		entry.LevelName = nil
		entry.Message = nil
		entry.Caller = nil
		clearMap(entry.Fields)
		clearFloatMap(entry.CustomMetrics)
		entry.Tags = clearByteSliceSlice(entry.Tags)
		entry.PID = 0
		entry.GoroutineID = nil
		entry.TraceID = nil
		entry.SpanID = nil
		entry.UserID = nil
		entry.SessionID = nil
		entry.RequestID = nil
		entry.Duration = 0
		entry.Error = nil
		entry.StackTrace = nil
		entry.Hostname = nil
		entry.Application = nil
		entry.Version = nil
		entry.Environment = nil

		// Update metrics
		globalEntryMetrics.IncEntryReused()
		return entry
	default:
		// Pool empty, create new entry from global pool
		entry := GetGlobalEntry()
		// globalEntryMetrics.poolMissCount.Add(1) // GetGlobalEntry already increments this
		return entry
	}
}

// PutEntryToPool returns a LogEntry to the pool
func PutEntryToPool(entry *LogEntry) {
	if entry.Caller != nil {
		PutCallerToPool(entry.Caller)
		entry.Caller = nil
	}
	// Return stack trace buffer to pool if it was used
	if entry.StackTraceBufPtr != nil {
		PutBuffer(entry.StackTraceBufPtr)
		entry.StackTraceBufPtr = nil
	}
	// Use goroutine-local pool if available
	localPool := GetGoroutineLocalEntryPool()
	localPool.PutLocalEntry(entry)
}

// ZeroAllocJSONSerialize serializes LogEntry to JSON without allocation
func (le *LogEntry) ZeroAllocJSONSerialize() []byte {
	// Get buffer from pool for zero allocation
	bufPtr := GetBuffer()
	buf := *bufPtr

	// Start with opening brace
	buf = append(buf, '{')

	// Serialisasi field-field penting
	buf = le.serializeTimestamp(buf, "timestamp", le.Timestamp)
	buf = append(buf, ',')

	// Serialize important fields - now LevelName is []byte
	buf = le.serializeByteSliceField(buf, "level", le.LevelName)
	buf = append(buf, ',')

	buf = le.serializeByteSliceField(buf, "message", le.Message)

	// Add other fields if any
	if le.PID != 0 {
		buf = append(buf, ',')
		buf = le.serializeIntField(buf, "pid", le.PID)
	}

	if le.GoroutineID != nil {
		buf = append(buf, ',')
		buf = le.serializeByteSliceField(buf, "goroutine_id", le.GoroutineID)
	}

	// Close with closing brace
	buf = append(buf, '}')

	// Save result and return buffer to pool
	result := make([]byte, len(buf))
	copy(result, buf)

	// Reset buffer and return to pool
	*bufPtr = (*bufPtr)[:0]
	PutBuffer(bufPtr)

	return result
}

// serializeTimestamp serializes a timestamp field
func (le *LogEntry) serializeTimestamp(buf []byte, key string, value time.Time) []byte {
	buf = append(buf, '"')
	buf = append(buf, key...)
	buf = append(buf, '"')
	buf = append(buf, ':')
	buf = append(buf, '"')

	// Format timestamp manually without allocation
	ts := value.Format(time.RFC3339)
	buf = append(buf, ts...)

	buf = append(buf, '"')
	return buf
}

// serializeByteSliceField serializes a byte slice field
func (le *LogEntry) serializeByteSliceField(buf []byte, key string, value []byte) []byte {
	buf = append(buf, '"')
	buf = append(buf, key...)
	buf = append(buf, '"')
	buf = append(buf, ':')
	buf = append(buf, '"')
	if value != nil {
		buf = append(buf, value...)
	} else {
		buf = append(buf, []byte("null")...)
	}
	buf = append(buf, '"')
	return buf
}

// serializeIntField serializes an int field
func (le *LogEntry) serializeIntField(buf []byte, key string, value int) []byte {
	buf = append(buf, '"')
	buf = append(buf, key...)
	buf = append(buf, '"')
	buf = append(buf, ':')

	// Convert int to string without allocation
	var temp [20]byte
	i := len(temp)
	val := value

	if val == 0 {
		buf = append(buf, '0')
		return buf
	}

	if val < 0 {
		buf = append(buf, '-')
		val = -val
	}

	for val > 0 && i > 0 {
		i--
		temp[i] = byte(val%10) + '0'
		val /= 10
	}

	return append(buf, temp[i:]...)
}

// ErrAppend is an optional interface that errors can implement to write
// their error message directly to a bytes.Buffer, avoiding intermediate string allocations.
type ErrAppend interface {
	AppendError(buf *bytes.Buffer)
}

// formatLogToBytes writes log data manually to byte buffer
func (le *LogEntry) formatLogToBytes(buf []byte) []byte {
	// Format: TIMESTAMP LEVEL MESSAGE [FIELDS] [TAGS]

	// Write timestamp
	ts := le.Timestamp.Format("2006-01-02T15:04:05.000Z07:00")
	buf = append(buf, ts...)
	buf = append(buf, ' ')

	// Write level - now LevelName is []byte
	buf = append(buf, le.LevelName...)
	buf = append(buf, ' ')

	// Write message
	buf = append(buf, le.Message...)
	buf = append(buf, ' ')

	// Write fields if any
	if len(le.Fields) > 0 {
		buf = append(buf, '[')
		first := true
		for k, v := range le.Fields {
			if !first {
				buf = append(buf, ',')
			}
			buf = append(buf, k...)
			buf = append(buf, ':')

			// Since v is []byte, we can directly append it
			buf = append(buf, v...)
			first = false
		}
		buf = append(buf, ']')
		buf = append(buf, ' ')
	}

	// Write tags if any
	if len(le.Tags) > 0 {
		buf = append(buf, '[')
		for i, tag := range le.Tags {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = append(buf, tag...)
		}
		buf = append(buf, ']')
	}

	return buf
}

// intToBytes writes integer to buffer without allocation
func (le *LogEntry) intToBytes(buf []byte, value int) []byte {
	if value == 0 {
		return append(buf, '0')
	}

	// For negative numbers
	if value < 0 {
		buf = append(buf, '-')
		value = -value
	}

	// Conversion without allocation
	var temp [20]byte
	i := len(temp)
	for value > 0 && i > 0 {
		i--
		temp[i] = byte(value%10) + '0'
		value /= 10
	}

	return append(buf, temp[i:]...)
}

// int64ToBytes writes int64 to buffer without allocation
func (le *LogEntry) int64ToBytes(buf []byte, value int64) []byte {
	if value == 0 {
		return append(buf, '0')
	}

	// For negative numbers
	if value < 0 {
		buf = append(buf, '-')
		value = -value
	}

	// Conversion without allocation
	var temp [20]byte
	i := len(temp)
	for value > 0 && i > 0 {
		i--
		temp[i] = byte(value%10) + '0'
		value /= 10
	}

	return append(buf, temp[i:]...)
}

// byteSlicePool for float formatting in core package
var byteSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 32) // Size sufficient for float formatting
	},
}

// GetByteSliceFromPool gets a byte slice from the pool
func GetByteSliceFromPool() []byte {
	return byteSlicePool.Get().([]byte)
}

// MaxSmallSlicePoolSize constant for compile-time configuration
const (
	MaxSmallSlicePoolSize = 1024 // 1KB limit for small slices in pool
	DefaultBufferSize     = MediumEntryBufferSize
)

// PutByteSliceToPool returns a byte slice to the pool
func PutByteSliceToPool(b []byte) {
	b = b[:0] // to
	//nolint:staticcheck
	byteSlicePool.Put(b)
}

// floatToBytes writes float64 to buffer without allocation
func (le *LogEntry) floatToBytes(buf []byte, value float64) []byte {
	// Use pooled buffer for float conversion
	tempBuf := GetByteSliceFromPool()
	defer PutByteSliceToPool(tempBuf)

	// Use strconv.AppendFloat for allocation-free conversion
	floatBytes := strconv.AppendFloat(tempBuf[:0], value, 'f', 2, 64)
	return append(buf, floatBytes...)
}

// PutLocalEntry returns entry to local goroutine pool
func (g *LocalPool) PutLocalEntry(entry *LogEntry) {
	if entry.Caller != nil {
		PutCallerToPool(entry.Caller)
		entry.Caller = nil
	}

	// Update metrics
	globalEntryMetrics.IncEntrySerialized()
	globalEntryMetrics.SetLastOperationTime(time.Now())

	// Try to put into local pool
	select {
	case g.entries <- entry:
		// Successfully put into pool
	default:
		// Pool full, return to global pool
		entryPool.Put(entry)
	}
}
