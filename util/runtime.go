package util

import (
	"path/filepath"
	runtime "runtime"
	"strings"
	"mire/core"
)

func GetCallerInfo(skip int) *core.CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	ci := core.GetCallerInfoFromPool()
	ci.File = filepath.Base(file)
	ci.Line = line
	
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		fullName := fn.Name()
		lastSlash := strings.LastIndex(fullName, "/")
		if lastSlash > 0 {
			pkgNameEnd := strings.Index(fullName[lastSlash+1:], ".")
			if pkgNameEnd > 0 {
				ci.Package = fullName[lastSlash+1 : lastSlash+1+pkgNameEnd]
				ci.Function = fullName[lastSlash+1+pkgNameEnd+1:]
			} else {
				ci.Function = fullName
			}
		} else {
			ci.Function = fullName
		}
	}
	
	return ci
}

// GetStackTrace returns a stack trace as a []byte slice from a pooled buffer,
// and the pointer to the pooled buffer. The caller is responsible for returning
// the buffer to the pool using core.PutBufferToPool (via the returned pointer).
func GetStackTrace(depth int) ([]byte, *[]byte) {
	bufPtr := core.GetBufferFromPool() // Get a pooled buffer
	tempBuf := *bufPtr // Dereference to get the actual []byte slice

	n := runtime.Stack(tempBuf, false)
	if n == 0 {
		core.PutBufferToPool(bufPtr) // Return empty buffer if no stack trace
		return nil, nil
	}
	
	// Trim to actual size used by runtime.Stack
	trace := tempBuf[:n]

	// Limit lines by depth. runtime.Stack output starts with "goroutine X [state]:\n"
	// and then has "package.function(...)\n\tfile:line +0xOFFSET\n" pairs.
	// We want to count pairs.
	lineCount := 0
	lastNewline := -1
	for i := 0; i < len(trace); i++ {
		if trace[i] == '\n' {
			lineCount++
			lastNewline = i
			// Each stack frame is typically two lines, plus initial goroutine line.
			// So 2*depth+1 lines for frames + 1 for goroutine header
			// Adjusting for the initial "goroutine" line
			if lineCount > (2 * depth + 1) { // 1 for goroutine header, 2 lines per frame
				trace = trace[:lastNewline]
				break
			}
		}
	}
	
	// The returned slice is part of the pooled buffer, and we return the pointer to it.
	// The caller will put the bufPtr back to the pool.
	return trace, bufPtr
}
