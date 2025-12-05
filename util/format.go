package util

import (
	"bytes"
	"strconv" // Re-added strconv import
	"time"

	"github.com/Lunar-Chipter/mire/core"
)

func FormatValue(buf *bytes.Buffer, value interface{}, maxWidth int) {
	// Use a pooled byte slice for intermediate conversions
	tempBuf := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(tempBuf)

	var needsQuote bool
	var content []byte // The content to write, potentially truncated

	switch v := value.(type) {
	case string:
		content = StringToBytes(v) // Zero-copy conversion - using function from util package
	case []byte:
		content = v
	case int:
		content = strconv.AppendInt(tempBuf[:0], int64(v), 10) // Reset length before appending
	case int8:
		content = strconv.AppendInt(tempBuf[:0], int64(v), 10) // Reset length before appending
	case int16:
		content = strconv.AppendInt(tempBuf[:0], int64(v), 10) // Reset length before appending
	case int32:
		content = strconv.AppendInt(tempBuf[:0], int64(v), 10) // Reset length before appending
	case int64:
		content = strconv.AppendInt(tempBuf[:0], v, 10) // Reset length before appending
	case uint:
		content = strconv.AppendUint(tempBuf[:0], uint64(v), 10) // Reset length before appending
	case uint8:
		content = strconv.AppendUint(tempBuf[:0], uint64(v), 10) // Reset length before appending
	case uint16:
		content = strconv.AppendUint(tempBuf[:0], uint64(v), 10) // Reset length before appending
	case uint32:
		content = strconv.AppendUint(tempBuf[:0], uint64(v), 10) // Reset length before appending
	case uint64:
		content = strconv.AppendUint(tempBuf[:0], v, 10) // Reset length before appending
	case float32:
		content = strconv.AppendFloat(tempBuf[:0], float64(v), 'f', 2, 32) // 'f' format, 2 decimal places, 32-bit float
	case float64:
		content = strconv.AppendFloat(tempBuf[:0], v, 'f', 2, 64) // 'f' format, 2 decimal places, 64-bit float
	case bool:
		content = strconv.AppendBool(tempBuf[:0], v) // Reset length before appending
	case error: // Handle errors directly
		if appender, ok := v.(core.ErrorAppender); ok { // Changed to core.ErrorAppender
			appender.AppendError(buf)
			return
		}
		content = StringToBytes(v.Error()) // Fallback, still allocates a string internally for v.Error()
	default:
		// Fallback for complex types - manual string conversion to avoid fmt
		tempStr := convertValueToString(v)
		content = StringToBytes(tempStr)
	}

	// Handle max width truncation
	if maxWidth > 0 && len(content) > maxWidth {
		buf.Write(content[:maxWidth])
		buf.Write([]byte("..."))
		return
	}

	// Determine if quoting is needed (value contains space)
	needsQuote = bytes.Contains(content, []byte(" "))

	if needsQuote {
		buf.WriteByte('"')
		buf.Write(content)
		buf.WriteByte('"')
	} else {
		buf.Write(content)
	}
}

func FormatTimestamp(buf *bytes.Buffer, t time.Time, format string) {
	// Use a pooled byte slice for intermediate conversions
	tempBuf := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(tempBuf)

	// Format the timestamp to the temp buffer
	tsBytes := t.AppendFormat(tempBuf[:0], format)
	buf.Write(tsBytes)
}

// convertValueToString manually converts common types to string without fmt
// Kept for compatibility - uses the public function
func convertValueToString(value interface{}) string {
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

// WriteInt writes an integer value directly to a buffer with minimal allocation
func WriteInt(buf *bytes.Buffer, value int64) {
	tempBuf := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(tempBuf)

	// Use AppendInt to format the integer
	bytes := strconv.AppendInt(tempBuf[:0], value, 10) // Reset length before appending
	buf.Write(bytes)
}

// WriteUint writes an unsigned integer value directly to a buffer with minimal allocation
func WriteUint(buf *bytes.Buffer, value uint64) {
	tempBuf := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(tempBuf)

	// Use AppendUint to format the unsigned integer
	bytes := strconv.AppendUint(tempBuf[:0], value, 10) // Reset length before appending
	buf.Write(bytes)
}

// WriteFloat writes a float value directly to a buffer with minimal allocation
func WriteFloat(buf *bytes.Buffer, value float64) {
	tempBuf := GetSmallByteSliceFromPool()
	defer PutSmallByteSliceToPool(tempBuf)

	// Use AppendFloat to format the float
	bytes := strconv.AppendFloat(tempBuf[:0], value, 'g', -1, 64) // Reset length before appending
	buf.Write(bytes)
}

// ConvertValue converts common types to string without fmt
func ConvertValue(value interface{}) string {
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
