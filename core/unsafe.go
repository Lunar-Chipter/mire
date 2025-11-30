package core

import (
	"unsafe"
)

// StringToBytes converts a string to a byte slice without memory allocation.
// WARNING: The returned byte slice shares memory with the string. It is read-only.
func StringToBytes(s string) (b []byte) {
	bh := (*[3]int)(unsafe.Pointer(&b))
	sh := (*[2]int)(unsafe.Pointer(&s))
	bh[0] = sh[0]
	bh[1] = sh[1]
	bh[2] = sh[1]
	return b
}

// BytesToString converts a byte slice to a string without memory allocation.
// WARNING: The returned string shares memory with the byte slice. 
// The byte slice should not be modified after this conversion.
func BytesToString(b []byte) (s string) {
	bh := (*[2]int)(unsafe.Pointer(&b))
	sh := (*[3]int)(unsafe.Pointer(&s))
	sh[0] = bh[0]
	sh[1] = bh[1]
	return
}