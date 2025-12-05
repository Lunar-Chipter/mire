package util

import (
	"bytes"
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

// BenchmarkFormatValue benchmarks the FormatValue function
func BenchmarkFormatValue(b *testing.B) {
	var buf bytes.Buffer
	values := []interface{}{
		"hello world",
		42,
		int64(123456),
		3.14159,
		true,
		false,
		nil,
		[]byte("byte slice"),
		[]int{1, 2, 3},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		FormatValue(&buf, values[i%len(values)], 0)
	}
}

// BenchmarkFormatValueWithMaxWidth benchmarks FormatValue with max width
func BenchmarkFormatValueWithMaxWidth(b *testing.B) {
	var buf bytes.Buffer
	value := "this is a very long string that will exceed the specified max width and need to be truncated"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		FormatValue(&buf, value, 20)
	}
}

// BenchmarkFormatTimestamp benchmarks the FormatTimestamp function
func BenchmarkFormatTimestamp(b *testing.B) {
	var buf bytes.Buffer
	timestamp := time.Now()
	format := "2006-01-02T15:04:05.000Z07:00" // Standard timestamp format

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		FormatTimestamp(&buf, timestamp, format)
	}
}

// BenchmarkConvertValue benchmarks the ConvertValue function
func BenchmarkConvertValue(b *testing.B) {
	values := []interface{}{
		"hello",
		42,
		int8(8),
		int16(16),
		int32(32),
		int64(64),
		uint(100),
		uint8(200),
		uint16(300),
		uint32(400),
		uint64(500),
		float32(3.14),
		float64(2.718),
		true,
		false,
		nil,
		[]byte("bytes"),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ConvertValue(values[i%len(values)])
	}
}

// BenchmarkWriteInt benchmarks the WriteInt function
func BenchmarkWriteInt(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteInt(&buf, int64(i))
	}
}

// BenchmarkWriteUint benchmarks the WriteUint function
func BenchmarkWriteUint(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteUint(&buf, uint64(i))
	}
}

// BenchmarkWriteFloat benchmarks the WriteFloat function
func BenchmarkWriteFloat(b *testing.B) {
	var buf bytes.Buffer
	value := 3.14159

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteFloat(&buf, value+float64(i)*0.001)
	}
}

// BenchmarkBufferPool benchmarks the buffer pool operations
func BenchmarkBufferPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := GetBufferFromPool()
		buf.Write([]byte("test data for buffer pool benchmark"))
		PutBufferToPool(buf)
	}
}

// BenchmarkSmallByteSlicePool benchmarks the small byte slice pool operations
func BenchmarkSmallByteSlicePool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := GetSmallByteSliceFromPool()
		slice = append(slice, []byte("test")...)
		PutSmallByteSliceToPool(slice)
	}
}

// BenchmarkStringToBytes benchmarks the StringToBytes conversion
func BenchmarkStringToBytes(b *testing.B) {
	str := "test string for StringToBytes benchmark"
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = core.StringToBytes(str)
	}
}

// BenchmarkBytesToString benchmarks the BytesToString conversion
func BenchmarkBytesToString(b *testing.B) {
	bytes := []byte("test bytes for BytesToString benchmark")
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BytesToString(bytes)
	}
}

// BenchmarkNow benchmarks the Now function with internal clock
func BenchmarkNow(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}

// BenchmarkClockNow benchmarks the clock Now function
func BenchmarkClockNow(b *testing.B) {
	clock := NewClock(10 * time.Millisecond)
	defer clock.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = clock.Now()
	}
}

// BenchmarkExtractFromContext benchmarks the ExtractFromContext function
func BenchmarkExtractFromContext(b *testing.B) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-12345")
	ctx = WithUserID(ctx, "user-67890")
	ctx = WithRequestID(ctx, "req-abcde")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		contextData := ExtractFromContext(ctx)
		PutMapStringToPool(contextData)
	}
}

// BenchmarkGetCallerInfo benchmarks the GetCallerInfo function
func BenchmarkGetCallerInfo(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		info := GetCallerInfo(2) // Get info for calling function
		if info != nil {
			core.PutCallerInfoToPool(info)
		}
	}
}

// BenchmarkGetStackTrace benchmarks the GetStackTrace function
func BenchmarkGetStackTrace(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stackTrace, bufPtr := GetStackTrace(10)
		if stackTrace != nil && bufPtr != nil {
			core.PutBufferToPool(bufPtr)
		}
	}
}

// BenchmarkGoroutineLocalBufferPool benchmarks the goroutine-local buffer pool
func BenchmarkGoroutineLocalBufferPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		localPool := GetGoroutineLocalBufferPool()
		buf := localPool.GetBufferFromLocalPool()
		if buf != nil {
			localPool.PutBufferToLocalPool(buf)
		} else {
			// Fallback to global pool if local is empty
			globalBuf := GetBufferFromPool()
			PutBufferToPool(globalBuf)
		}
	}
}

// BenchmarkMapStringPool benchmarks MapString pool operations
func BenchmarkMapStringPool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := GetMapStringFromPool()
		m["key"] = "value"
		m["key2"] = "value2"
		PutMapStringToPool(m)
	}
}

// BenchmarkStringSlicePool benchmarks StringSlice pool operations
func BenchmarkStringSlicePool(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := GetStringSliceFromPool()
		slice = append(slice, "item1", "item2", "item3")
		PutStringSliceToPool(slice)
	}
}

// BenchmarkRuntimeNumGoroutine benchmarks runtime.NumGoroutine call
func BenchmarkRuntimeNumGoroutine(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = runtime.NumGoroutine()
	}
}

// BenchmarkTimeNow benchmarks time.Now call for comparison
func BenchmarkTimeNow(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}