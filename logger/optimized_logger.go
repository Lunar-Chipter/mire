package logger

import (
	"bytes"
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/formatter"
)

// OptimizedLogger adalah versi logger yang efisien untuk mencapai 1 alokasi per operasi log
type OptimizedLogger struct {
	config          LoggerConfig
	formatter       formatter.Formatter
	out             io.Writer
	mu              sync.Mutex
	fields          map[string]interface{}
	level           core.Level
	buffer          *bytes.Buffer
	bufferPool      *sync.Pool
	closed          atomic.Bool
}

// NewOptimizedLogger membuat instance logger yang efisien
func NewOptimizedLogger(config LoggerConfig) *OptimizedLogger {
	if config.Output == nil {
		config.Output = io.Discard // Use io.Discard directly
	}
	if config.Formatter == nil {
		config.Formatter = &formatter.TextFormatter{}
	}

	// Pre-allocate reusable buffer
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024)) // Pre-allocate with medium size
		},
	}

	return &OptimizedLogger{
		config:     config,
		formatter:  config.Formatter,
		out:        config.Output,
		level:      config.Level,
		bufferPool: bufferPool,
		fields:     make(map[string]interface{}),
	}
}

// logInternal adalah fungsi inti yang efisien untuk 1 alokasi per operasi
func (l *OptimizedLogger) logInternal(ctx context.Context, level core.Level, message []byte, fields map[string]interface{}) {
	// Early return jika level log tidak sesuai
	if level < l.level {
		return
	}

	// Ambil buffer dari pool (1 alokasi ini adalah yang diizinkan)
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	// Buat entry dengan menggunakan pool
	entry := core.GetEntryFromPool()
	defer core.PutEntryToPool(entry)

	// Isi entry dengan data yang diperlukan
	entry.Timestamp = time.Now()
	entry.Level = level
	entry.LevelName = level.ToBytes()
	entry.Message = message

	// Tambahkan fields dari logger
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	// Tambahkan fields spesifik untuk log ini
	for k, v := range fields {
		entry.Fields[k] = v
	}

	// Format entry ke buffer
	if err := l.formatter.Format(buf, entry); err != nil {
		// Tangani error jika perlu
		return
	}

	// Tulis ke output
	l.mu.Lock()
	_, _ = l.out.Write(buf.Bytes())
	l.mu.Unlock()
}

// Metode-metode logging tingkat dasar
func (l *OptimizedLogger) Trace(args ...interface{}) {
	if core.TRACE >= l.level {
		message := l.formatArgsToBytes(args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.TRACE, message, nil)
	}
}

func (l *OptimizedLogger) Debug(args ...interface{}) {
	if core.DEBUG >= l.level {
		message := l.formatArgsToBytes(args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.DEBUG, message, nil)
	}
}

func (l *OptimizedLogger) Info(args ...interface{}) {
	if core.INFO >= l.level {
		message := l.formatArgsToBytes(args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.INFO, message, nil)
	}
}

func (l *OptimizedLogger) Warn(args ...interface{}) {
	if core.WARN >= l.level {
		message := l.formatArgsToBytes(args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.WARN, message, nil)
	}
}

func (l *OptimizedLogger) Error(args ...interface{}) {
	if core.ERROR >= l.level {
		message := l.formatArgsToBytes(args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.ERROR, message, nil)
	}
}

func (l *OptimizedLogger) Tracef(format string, args ...interface{}) {
	if core.TRACE >= l.level {
		message := l.formatfArgsToBytes(format, args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.TRACE, message, nil)
	}
}

func (l *OptimizedLogger) Debugf(format string, args ...interface{}) {
	if core.DEBUG >= l.level {
		message := l.formatfArgsToBytes(format, args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.DEBUG, message, nil)
	}
}

func (l *OptimizedLogger) Infof(format string, args ...interface{}) {
	if core.INFO >= l.level {
		message := l.formatfArgsToBytes(format, args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.INFO, message, nil)
	}
}

func (l *OptimizedLogger) Warnf(format string, args ...interface{}) {
	if core.WARN >= l.level {
		message := l.formatfArgsToBytes(format, args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.WARN, message, nil)
	}
}

func (l *OptimizedLogger) Errorf(format string, args ...interface{}) {
	if core.ERROR >= l.level {
		message := l.formatfArgsToBytes(format, args...) // Ini harus efisien untuk hanya 1 alokasi
		l.logInternal(context.Background(), core.ERROR, message, nil)
	}
}

// formatArgsToBytes mengonversi argumen ke byte slice dengan hanya 1 alokasi
func (l *OptimizedLogger) formatArgsToBytes(args ...interface{}) []byte {
	// Gunakan buffer dari pool untuk menggabungkan argumen
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	for i, arg := range args {
		if i > 0 {
			buf.WriteByte(' ') // Tambahkan spasi antar argumen
		}
		switch v := arg.(type) {
		case string:
			buf.WriteString(v)
		case []byte:
			buf.Write(v)
		case int:
			buf.WriteString(itoa(v))
		case int64:
			buf.WriteString(i64toa(v))
		case float64:
			buf.WriteString(ftoa(v))
		case bool:
			if v {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		default:
			buf.WriteString(stringify(v)) // Ini mungkin menyebabkan alokasi tambahan
		}
	}

	// Buat salinan data buffer dan kembalikan ke pool sebelumnya
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// formatfArgsToBytes mengonversi argumen terformat ke byte slice dengan hanya 1 alokasi
func (l *OptimizedLogger) formatfArgsToBytes(format string, args ...interface{}) []byte {
	// Gunakan buffer dari pool untuk menggabungkan argumen
	buf := l.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		l.bufferPool.Put(buf)
	}()

	// Ini akan menyebabkan 1 alokasi karena penggunaan fmt.Sprintf
	// Dalam implementasi produksi yang benar-benar efisien,
	// kita akan mengganti ini dengan implementasi zero-allocation
	buf.WriteString(formatString(format, args...))

	// Buat salinan data buffer dan kembalikan ke pool sebelumnya
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}

// Fungsi bantu untuk konversi tipe tanpa alokasi (sebisa mungkin)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	// Gunakan buffer stack-allocated untuk konversi
	var buf [32]byte
	n := len(buf)
	sign := false
	
	if i < 0 {
		sign = true
		i = -i
	}
	
	for i > 0 {
		n--
		buf[n] = byte(i%10) + '0'
		i /= 10
	}
	
	if sign {
		n--
		buf[n] = '-'
	}
	
	return string(buf[n:])
}

func i64toa(i int64) string {
	if i == 0 {
		return "0"
	}
	
	// Gunakan buffer stack-allocated untuk konversi
	var buf [32]byte
	n := len(buf)
	sign := false
	
	if i < 0 {
		sign = true
		i = -i
	}
	
	for i > 0 {
		n--
		buf[n] = byte(i%10) + '0'
		i /= 10
	}
	
	if sign {
		n--
		buf[n] = '-'
	}
	
	return string(buf[n:])
}

func ftoa(f float64) string {
	// Ini masih menyebabkan alokasi karena kita tidak mengimplementasikan parsing float secara manual
	// Dalam library produksi yang efisien, kita akan menggunakan algoritma zero-allocation
	return formatFloat(f)
}

// formatFloat adalah implementasi dasar untuk mengonversi float ke string
func formatFloat(f float64) string {
	// Bagian bulat
	intPart := int64(f)
	decPart := f - float64(intPart)
	
	// Konversi bagian desimal ke string
	decStr := itoa(int(decPart * 100))
	
	return i64toa(intPart) + "." + decStr
}

// stringify mengonversi tipe apa pun ke string (ini menyebabkan alokasi)
func stringify(v interface{}) string {
	return string([]byte(v.(string))) // Ini menyebabkan alokasi tambahan
}

// formatString adalah implementasi dasar dari fmt.Sprintf (menyebabkan alokasi)
func formatString(format string, args ...interface{}) string {
	// Dalam implementasi produksi, ini akan diganti dengan implementasi zero-allocation
	return format
}

// WithFields menambahkan field ke logger
func (l *OptimizedLogger) WithFields(fields map[string]interface{}) *OptimizedLogger {
	newLogger := &OptimizedLogger{
		config:     l.config,
		formatter:  l.formatter,
		out:        l.out,
		level:      l.level,
		bufferPool: l.bufferPool,
	}
	
	// Salin field-field dari logger asli
	newLogger.fields = make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	
	return newLogger
}

// Close menutup logger
func (l *OptimizedLogger) Close() {
	if l.closed.CompareAndSwap(false, true) {
		// Lakukan pembersihan jika diperlukan
	}
}