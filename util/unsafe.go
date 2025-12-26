//go:build !no_colors && !no_metrics

package util

import (
	"errors"
	"sync"
	"unsafe"
)

type ZeroAllocBuffer struct {
	buf []byte
	len int
	_   [64 - unsafe.Sizeof(int(0))]byte // Padding for cache alignment
}

// Error definitions
var (
	ErrBufferOverflow = errors.New("buffer overflow")
)

func (b *ZeroAllocBuffer) WriteBytes(data []byte) error {
	if b.available() < len(data) {
		return ErrBufferOverflow
	}

	copy(b.buf[b.len:], data)
	b.len += len(data)
	return nil
}

// WriteByte writes a single byte to the buffer
func (b *ZeroAllocBuffer) WriteByte(c byte) error {
	if b.len >= len(b.buf) {
		return ErrBufferOverflow
	}
	b.buf[b.len] = c
	b.len++
	return nil
}

// WriteString writes a string to the buffer
func (b *ZeroAllocBuffer) WriteString(s string) error {
	return b.WriteBytes([]byte(s))
}

// available returns the available space in the buffer
func (b *ZeroAllocBuffer) available() int {
	return len(b.buf) - b.len
}

// Bytes returns the buffer content
func (b *ZeroAllocBuffer) Bytes() []byte {
	return b.buf[:b.len]
}

// Len returns the length of the buffer content
func (b *ZeroAllocBuffer) Len() int {
	return b.len
}

// Reset resets the buffer
func (b *ZeroAllocBuffer) Reset() {
	b.len = 0
}

// ZeroAllocBufferPool is a pool of ZeroAllocBuffer instances
var ZeroAllocBufferPool = sync.Pool{
	New: func() interface{} {
		return &ZeroAllocBuffer{
			buf: make([]byte, 0, MediumBufferSize), // Pre-allocated
		}
	},
}

// GetZeroAllocBuffer gets a ZeroAllocBuffer from the pool
func GetZeroAllocBuffer() *ZeroAllocBuffer {
	buf := ZeroAllocBufferPool.Get().(*ZeroAllocBuffer)
	buf.Reset()
	return buf
}

// PutZeroAllocBuffer returns a ZeroAllocBuffer to the pool
func PutZeroAllocBuffer(buf *ZeroAllocBuffer) {
	ZeroAllocBufferPool.Put(buf)
}

// ColorByteSlice represents pre-allocated color byte slices
var (
	ErrorColor = []byte("\x1b[38;5;196m")
	WarnColor  = []byte("\x1b[38;5;220m")
	InfoColor  = []byte("\x1b[38;5;75m")
	DebugColor = []byte("\x1b[38;5;245m")
	ResetColor = []byte("\x1b[0m")
)

// StringToBytes converts string to []byte without allocation (shares memory)
//
// CRITICAL SAFETY WARNING:
// This function uses unsafe operations to convert string to []byte without copying memory.
// The returned byte slice SHARES the underlying memory with the original string.
//
// IMPORTANT USAGE RULES:
// 1. DO NOT modify the returned byte slice - this will corrupt the original string
// 2. DO NOT append to the returned byte slice
// 3. DO NOT store the returned byte slice longer than the original string
// 4. For data that needs modification, always create a copy: b := []byte(s)
// 5. This is safe only for read-only access or when you control the lifetime
//
// This is a zero-allocation optimization. Only use this when you fully understand
// the implications and follow the safety rules above.
func StringToBytes(s string) (b []byte) {
	bh := (*[3]int)(unsafe.Pointer(&b))
	sh := (*[2]int)(unsafe.Pointer(&s))
	bh[0] = sh[0]
	bh[1] = sh[1]
	bh[2] = sh[1]
	return b
}

// B2s converts []byte to string without allocation (shares memory)
//
// CRITICAL SAFETY WARNING:
// This function uses unsafe operations to convert []byte to string without copying memory.
// The returned string SHARES underlying memory with the original []byte.
//
// IMPORTANT USAGE RULES:
// 1. DO NOT modify the original []byte after calling this function
// 2. DO NOT append to the original []byte after calling this function
// 3. DO NOT store the returned string longer than the original []byte
// 4. Only use this when you control the lifetime of both values
//
// This is a zero-allocation optimization. Only use when you fully understand
// implications and follow the safety rules above.
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// BytesToString converts []byte to string
func BytesToString(b []byte) string {
	return B2s(b)
}
