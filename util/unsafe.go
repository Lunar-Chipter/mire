//go:build !no_colors && !no_metrics

package util

import (
	"errors"
	"sync"
	"unsafe"
)

// ZeroAllocBuffer represents a zero-allocation buffer for high-performance logging
// Prinsip: "No hidden costs, no runtime surprises"
type ZeroAllocBuffer struct {
	buf []byte
	len int
	_   [64 - unsafe.Sizeof(int(0))]byte // Cache line alignment
}

// Error definitions
var (
	ErrBufferOverflow = errors.New("buffer overflow")
)

// WriteBytes writes bytes to the buffer with manual byte manipulation
// Prinsip: "Control over every byte"
//go:inline
func (b *ZeroAllocBuffer) WriteBytes(data []byte) error {
	if b.available() < len(data) {
		return ErrBufferOverflow
	}
	
	// Manual byte copying untuk performa maksimal
	copy(b.buf[b.len:], data)
	b.len += len(data)
	return nil
}

// WriteByte writes a single byte to the buffer with O(1) performance
// Prinsip: "Predictable performance under all conditions"
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
// Prinsip: "All pre-allocated and reusable"
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
// Prinsip: "Beautiful logs without performance tax"
var (
	ErrorColor   = []byte("\x1b[38;5;196m")
	WarnColor    = []byte("\x1b[38;5;220m")
	InfoColor    = []byte("\x1b[38;5;75m")
	DebugColor   = []byte("\x1b[38;5;245m")
	ResetColor   = []byte("\x1b[0m")
)

// S2b converts a string to a byte slice without memory allocation.
// WARNING: The returned byte slice shares memory with the string. It is read-only.
func S2b(s string) (b []byte) {
	bh := (*[3]int)(unsafe.Pointer(&b))
	sh := (*[2]int)(unsafe.Pointer(&s))
	bh[0] = sh[0]
	bh[1] = sh[1]
	bh[2] = sh[1]
	return b
}

// B2s converts byte slice to a string without memory allocation.
// WARNING: The returned string shares memory with the byte slice. Do not modify the bytes.
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}