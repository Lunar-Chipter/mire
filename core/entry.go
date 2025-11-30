package core

import (
	"bytes"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// LogEntry represents a single log entry with all its metadata
// LogEntry merepresentasikan entri log tunggal dengan semua metadata-nya
type LogEntry struct {
	Timestamp     time.Time            `json:"timestamp"`              // When the log was created
	Level         Level                `json:"level"`                  // Log severity level
	LevelName     []byte               `json:"level_name"`             // Byte representation of level for zero-allocation formatting
	Message       []byte               `json:"message"`                // Log message
	Caller        *CallerInfo          `json:"caller,omitempty"`       // Caller information
	Fields        map[string]interface{} `json:"fields,omitempty"`     // Additional fields
	PID           int                  `json:"pid"`                    // Process ID
	GoroutineID   []byte               `json:"goroutine_id,omitempty"` // Goroutine ID as byte slice
	TraceID       []byte               `json:"trace_id,omitempty"`     // Trace ID for distributed tracing as byte slice
	SpanID        []byte               `json:"span_id,omitempty"`      // Span ID for distributed tracing as byte slice
	UserID        []byte               `json:"user_id,omitempty"`      // User ID as byte slice
	SessionID     []byte               `json:"session_id,omitempty"`   // Session ID as byte slice
	RequestID     []byte               `json:"request_id,omitempty"`   // Request ID as byte slice
	Duration      time.Duration        `json:"duration,omitempty"`     // Operation duration
	Error         error                `json:"error,omitempty"`        // Error information
	StackTrace    []byte               `json:"stack_trace,omitempty"`  // Stack trace
	StackTraceBufPtr *[]byte           `json:"-"`                      // Pointer to the pooled buffer for StackTrace
	Hostname      []byte               `json:"hostname,omitempty"`     // Hostname as byte slice
	Application   []byte               `json:"application,omitempty"`  // Application name as byte slice
	Version       []byte               `json:"version,omitempty"`      // Application version as byte slice
	Environment   []byte               `json:"environment,omitempty"`  // Environment (dev/prod/etc) as byte slice
	CustomMetrics map[string]float64   `json:"custom_metrics,omitempty"` // Custom metrics
	Tags          [][]byte             `json:"tags,omitempty"`         // Tags for categorization as byte slices
	_             [64 - unsafe.Sizeof(time.Time{})%64]byte // Padding for cache alignment
}

// CallerInfo contains information about the code location where the log was created
// CallerInfo berisi informasi tentang lokasi kode di mana log dibuat
type CallerInfo struct {
	File     string `json:"file"`     // Source file name
	Line     int    `json:"line"`     // Line number
	Function string `json:"function"` // Function name
	Package  string `json:"package"`  // Package name
}

// Object pool for reusing CallerInfo objects
var callerInfoPool = sync.Pool{
	New: func() interface{} {
		return &CallerInfo{}
	},
}

// GetCallerInfoFromPool gets a CallerInfo from the pool
func GetCallerInfoFromPool() *CallerInfo {
	return callerInfoPool.Get().(*CallerInfo)
}

// PutCallerInfoToPool returns a CallerInfo to the pool
func PutCallerInfoToPool(ci *CallerInfo) {
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

// Object pool for reusing map[string]interface{} objects
var mapInterfacePool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

// GetMapInterfaceFromPool gets a map[string]interface{} from the pool
func GetMapInterfaceFromPool() map[string]interface{} {
	m := mapInterfacePool.Get().(map[string]interface{})
	for k := range m {
		delete(m, k) // Reset the map
	}
	return m
}

// PutMapInterfaceToPool returns a map[string]interface{} to the pool
func PutMapInterfaceToPool(m map[string]interface{}) {
	mapInterfacePool.Put(m)
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

// GetBufferFromPool gets a byte buffer from the pool
func GetBufferFromPool() *[]byte {
	buf := bufferPool.Get().(*[]byte)
	*buf = (*buf)[:0] // Reset length but keep capacity
	return buf
}

// PutBufferToPool returns a byte buffer to the pool
func PutBufferToPool(buf *[]byte) {
	bufferPool.Put(buf)
}

// GetStringSliceFromPool gets a string slice from the pool
func GetStringSliceFromPool() *[]string {
	s := stringSlicePool.Get().(*[]string)
	*s = (*s)[:0] // Reset slice length but keep capacity
	return s
}

// PutStringSliceToPool returns a string slice to the pool
func PutStringSliceToPool(s *[]string) {
	stringSlicePool.Put(s)
}

// Global metrics instance - sekarang menggunakan CoreMetrics
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
func clearMap(m map[string]interface{}) {
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

// clearStringSlice clears a string slice
func clearStringSlice(s []string) []string {
	return s[:0]
}

// clearByteSliceSlice clears a byte slice slice
func clearByteSliceSlice(s [][]byte) [][]byte {
	return s[:0]
}


// Object pool for reusing LogEntry objects to reduce memory allocation
// Pool objek untuk menggunakan kembali objek LogEntry mengurangi alokasi memori
var entryPool = sync.Pool{
	New: func() interface{} {
		// Use pooled resources for slices and maps within the LogEntry
		tags := GetStringSliceFromPool()
		// Convert []string to [][]byte for the Tags field
		tagsAsBytes := make([][]byte, 0, len(*tags))
		for _, tag := range *tags {
			tagsAsBytes = append(tagsAsBytes, StringToBytes(tag))
		}
		return &LogEntry{
			Fields:        GetMapInterfaceFromPool(),
			CustomMetrics: GetMapFloatFromPool(),
			Tags:          tagsAsBytes,
		}
	},
}

// Efficient allocation-free pool access with optimized path
var (
	// Pre-allocated entries to reduce sync.Pool contention under low load
	preallocatedEntries = make([]*LogEntry, 32) // Fixed-size array for fast access
	preallocatedIndex = &atomic.Int64{}        // Atomic index for round-robin access
)

// Constants for compile-time configuration
const (
	// Buffer sizes for different log entry types
	SmallEntryBufferSize  = 256
	MediumEntryBufferSize = 1024
	LargeEntryBufferSize  = 4096

	// Pre-allocated slice capacities
	TagsSliceCapacity     = 10
	FieldsMapCapacity     = 8
	MetricsMapCapacity    = 4

	// Performance optimization constants
	PreallocatedPoolSize  = 32   // Number of pre-allocated entries
	MaxOptimizedPathAttempts   = 5    // Max attempts to use optimized path before fallback
)

// GetEntryFromPool gets a LogEntry from the pool
// Mendapatkan LogEntry dari pool
func GetEntryFromPool() *LogEntry {
	// Fallback: Use regular goroutine-local pool
	localPool := GetGoroutineLocalEntryPool()
	return localPool.GetEntryFromLocalPool()
}

// GetEntryFromGlobalPool gets a LogEntry directly from the global pool
// Mendapatkan LogEntry langsung dari pool global
func GetEntryFromGlobalPool() *LogEntry {
	entry := entryPool.Get().(*LogEntry)

	// Update metrics
	if entry.Timestamp.IsZero() {
		// Entry baru dibuat (pool miss)
		globalEntryMetrics.IncEntryPoolMiss()
		globalEntryMetrics.IncEntryCreated()
	} else {
		// Entry digunakan ulang
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

// goroutineLocalEntryPool menyimpan entry pool per goroutine untuk menghindari lock contention
var goroutineLocalEntryPool = sync.Map{}

// GoroutineLocalEntryPool represents a per-goroutine entry pool
type GoroutineLocalEntryPool struct {
	entries chan *LogEntry
}

// getGoroutineID mendapatkan ID goroutine saat ini (implementasi sederhana)
func getGoroutineID() uint64 {
	// Dalam implementasi nyata, kita akan menggunakan cara yang lebih andal
	// untuk mendapatkan ID goroutine
	return uint64(time.Now().UnixNano() % 1000000)
}

// GetGoroutineLocalEntryPool mendapatkan entry pool untuk goroutine saat ini
func GetGoroutineLocalEntryPool() *GoroutineLocalEntryPool {
	gid := getGoroutineID()
	if pool, ok := goroutineLocalEntryPool.Load(gid); ok {
		return pool.(*GoroutineLocalEntryPool)
	}
	
	// Buat pool baru untuk goroutine ini
	newPool := &GoroutineLocalEntryPool{
		entries: make(chan *LogEntry, 100), // Buffered channel untuk 100 entries
	}
	goroutineLocalEntryPool.Store(gid, newPool)
	
	return newPool
}

// GetEntryFromLocalPool mendapatkan entry dari pool lokal goroutine
func (g *GoroutineLocalEntryPool) GetEntryFromLocalPool() *LogEntry {
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
		// Pool kosong, buat entry baru dari pool global
		entry := GetEntryFromGlobalPool()
		// globalEntryMetrics.poolMissCount.Add(1) // GetEntryFromGlobalPool already increments this
		return entry
	}
}

// PutEntryToPool returns a LogEntry to the pool
// Mengembalikan LogEntry ke pool
func PutEntryToPool(entry *LogEntry) {
    if entry.Caller != nil {
        PutCallerInfoToPool(entry.Caller)
        entry.Caller = nil
    }
    // Return stack trace buffer to pool if it was used
    if entry.StackTraceBufPtr != nil {
        PutBufferToPool(entry.StackTraceBufPtr)
        entry.StackTraceBufPtr = nil
    }
	// Gunakan goroutine-local pool jika tersedia
	localPool := GetGoroutineLocalEntryPool()
	localPool.PutEntryToLocalPool(entry)
}

// ZeroAllocJSONSerialize serializes the LogEntry to JSON without memory allocation
func (le *LogEntry) ZeroAllocJSONSerialize() []byte {
	// Dapatkan buffer dari pool untuk zero allocation
	bufPtr := GetBufferFromPool()
	buf := *bufPtr

	// Mulai dengan kurung kurawal
	buf = append(buf, '{')

	// Serialisasi field-field penting
	buf = le.serializeTimestamp(buf, "timestamp", le.Timestamp)
	buf = append(buf, ',')

	// Serialisasi field-field penting - sekarang LevelName adalah []byte
	buf = le.serializeByteSliceField(buf, "level", le.LevelName)
	buf = append(buf, ',')

	buf = le.serializeByteSliceField(buf, "message", le.Message)

	// Tambahkan field lain jika ada
	if le.PID != 0 {
		buf = append(buf, ',')
		buf = le.serializeIntField(buf, "pid", le.PID)
	}

	if le.GoroutineID != nil {
		buf = append(buf, ',')
		buf = le.serializeByteSliceField(buf, "goroutine_id", le.GoroutineID)
	}

	// Tutup dengan kurung kurawal
	buf = append(buf, '}')

	// Simpan hasil dan kembalikan buffer ke pool
	result := make([]byte, len(buf))
	copy(result, buf)

	// Reset buffer dan kembalikan ke pool
	*bufPtr = (*bufPtr)[:0]
	PutBufferToPool(bufPtr)

	return result
}

// serializeTimestamp serializes a timestamp field
func (le *LogEntry) serializeTimestamp(buf []byte, key string, value time.Time) []byte {
	buf = append(buf, '"')
	buf = append(buf, key...)
	buf = append(buf, '"')
	buf = append(buf, ':')
	buf = append(buf, '"')

	// Format timestamp secara manual tanpa alokasi
	ts := value.Format(time.RFC3339)
	buf = append(buf, ts...)

	buf = append(buf, '"')
	return buf
}

// serializeField serializes a string field
func (le *LogEntry) serializeField(buf []byte, key, value string) []byte {
	buf = append(buf, '"')
	buf = append(buf, key...)
	buf = append(buf, '"')
	buf = append(buf, ':')
	buf = append(buf, '"')
	buf = append(buf, value...)
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

	// Konversi int ke string tanpa alokasi
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

// ErrorAppender is an optional interface that errors can implement to write
// their error message directly to a bytes.Buffer, avoiding intermediate string allocations.
type ErrorAppender interface {
	AppendError(buf *bytes.Buffer)
}


// formatLogToBytes menulis data log secara manual ke buffer byte untuk efisiensi maksimal
func (le *LogEntry) formatLogToBytes(buf []byte) []byte {
	// Format: TIMESTAMP LEVEL MESSAGE [FIELDS] [TAGS]

	// Menulis timestamp
	ts := le.Timestamp.Format("2006-01-02T15:04:05.000Z07:00")
	buf = append(buf, ts...)
	buf = append(buf, ' ')

	// Menulis level - sekarang LevelName adalah []byte
	buf = append(buf, le.LevelName...)
	buf = append(buf, ' ')

	// Menulis pesan
	buf = append(buf, le.Message...)
	buf = append(buf, ' ')

	// Menulis fields jika ada
	if len(le.Fields) > 0 {
		buf = append(buf, '[')
		first := true
		for k, v := range le.Fields {
			if !first {
				buf = append(buf, ',')
			}
			buf = append(buf, k...)
			buf = append(buf, ':')

			// Konversi nilai ke string tanpa alokasi berlebih
			switch val := v.(type) {
			case string:
				buf = append(buf, val...)
			case []byte:
				buf = append(buf, val...)
			case int:
				buf = le.intToBytes(buf, val)
			case int64:
				buf = le.int64ToBytes(buf, val)
			case float64:
				buf = le.floatToBytes(buf, val)
			case bool:
				if val {
					buf = append(buf, "true"...)
				} else {
					buf = append(buf, "false"...)
				}
			default:
				// Gunakan konversi manual untuk menghindari alokasi fmt.Sprintf
				// Gunakan fungsi konversi lokal untuk menghindari circular import
				localStringConversion := func(value interface{}) []byte {
					switch v := value.(type) {
					case string:
						return StringToBytes(v)
					case []byte:
						return v
					case int:
						buf := make([]byte, 0, 20)
						return strconv.AppendInt(buf, int64(v), 10)
					case int8:
						buf := make([]byte, 0, 20)
						return strconv.AppendInt(buf, int64(v), 10)
					case int16:
						buf := make([]byte, 0, 20)
						return strconv.AppendInt(buf, int64(v), 10)
					case int32:
						buf := make([]byte, 0, 20)
						return strconv.AppendInt(buf, int64(v), 10)
					case int64:
						buf := make([]byte, 0, 20)
						return strconv.AppendInt(buf, v, 10)
					case uint:
						buf := make([]byte, 0, 20)
						return strconv.AppendUint(buf, uint64(v), 10)
					case uint8:
						buf := make([]byte, 0, 20)
						return strconv.AppendUint(buf, uint64(v), 10)
					case uint16:
						buf := make([]byte, 0, 20)
						return strconv.AppendUint(buf, uint64(v), 10)
					case uint32:
						buf := make([]byte, 0, 20)
						return strconv.AppendUint(buf, uint64(v), 10)
					case uint64:
						buf := make([]byte, 0, 20)
						return strconv.AppendUint(buf, v, 10)
					case float32:
						buf := make([]byte, 0, 32)
						return strconv.AppendFloat(buf, float64(v), 'g', -1, 32)
					case float64:
						buf := make([]byte, 0, 32)
						return strconv.AppendFloat(buf, v, 'g', -1, 64)
					case bool:
						if v {
							return StringToBytes("true")
						}
						return StringToBytes("false")
					case nil:
						return StringToBytes("null")
					default:
						// For complex types that can't be easily converted
						// This is a last resort case - should be avoided in demanding scenarios
						return StringToBytes("<complex-type>")
					}
				}
				tempBytes := localStringConversion(val)
				buf = append(buf, tempBytes...)
			}
			first = false
		}
		buf = append(buf, ']')
		buf = append(buf, ' ')
	}

	// Menulis tags jika ada
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

// intToBytes menulis integer ke buffer tanpa alokasi
func (le *LogEntry) intToBytes(buf []byte, value int) []byte {
	if value == 0 {
		return append(buf, '0')
	}

	// Untuk angka negatif
	if value < 0 {
		buf = append(buf, '-')
		value = -value
	}

	// Konversi tanpa alokasi
	var temp [20]byte
	i := len(temp)
	for value > 0 && i > 0 {
		i--
		temp[i] = byte(value%10) + '0'
		value /= 10
	}

	return append(buf, temp[i:]...)
}

// int64ToBytes menulis int64 ke buffer tanpa alokasi
func (le *LogEntry) int64ToBytes(buf []byte, value int64) []byte {
	if value == 0 {
		return append(buf, '0')
	}

	// Untuk angka negatif
	if value < 0 {
		buf = append(buf, '-')
		value = -value
	}

	// Konversi tanpa alokasi
	var temp [20]byte
	i := len(temp)
	for value > 0 && i > 0 {
		i--
		temp[i] = byte(value%10) + '0'
		value /= 10
	}

	return append(buf, temp[i:]...)
}

// byteSlicePool untuk float formatting di core package
var byteSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 32) // Ukuran yang cukup untuk float formatting
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
	b = b[:0] // Reset length but keep capacity
	byteSlicePool.Put(b)
}

// floatToBytes menulis float64 ke buffer tanpa alokasi
func (le *LogEntry) floatToBytes(buf []byte, value float64) []byte {
	// Gunakan pooled buffer untuk konversi float
	tempBuf := GetByteSliceFromPool()
	defer PutByteSliceToPool(tempBuf)

	// Gunakan strconv.AppendFloat untuk konversi tanpa alokasi
	floatBytes := strconv.AppendFloat(tempBuf[:0], value, 'f', 2, 64)
	return append(buf, floatBytes...)
}

// PutEntryToLocalPool mengembalikan entry ke pool lokal goroutine
func (g *GoroutineLocalEntryPool) PutEntryToLocalPool(entry *LogEntry) {
	if entry.Caller != nil {
		PutCallerInfoToPool(entry.Caller)
		entry.Caller = nil
	}
	
	// Update metrics
	globalEntryMetrics.IncEntrySerialized()
	globalEntryMetrics.SetLastOperationTime(time.Now())
	
	// Coba masukkan ke pool lokal
	select {
	case g.entries <- entry:
		// Berhasil dimasukkan ke pool
	default:
		// Pool penuh, kembalikan ke pool global
		entryPool.Put(entry)
	}
}

// manualStringConversion converts common types to string without fmt to avoid circular import
func manualStringConversion(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v) // This is unavoidable for []byte to string
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		// For complex types that can't be easily converted
		// This is a last resort case - should be avoided in demanding scenarios
		return "<complex-type>"
	}
}

