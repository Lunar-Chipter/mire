package util

import (
	"github.com/Lunar-Chipter/mire/core"
	"path/filepath"
	"runtime"
	"strings"
)

func GetCallerInfo(skip int) *core.Caller {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	ci := core.GetCallerFromPool()
	ci.File = filepath.Base(file)
	ci.Line = line

	fn := runtime.FuncForPC(pc)
	var fullName string
	if fn != nil {
		fullName = fn.Name()
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
// and pointer to the pooled buffer. The caller is responsible for returning
// buffer to pool using core.PutBuffer (via returned pointer).
func GetStackTrace(depth int) ([]byte, *[]byte) {
	bufPtr := core.GetBuffer()
	tempBuf := *bufPtr

	n := runtime.Stack(tempBuf, false)
	if n == 0 {
		core.PutBuffer(bufPtr)
		return nil, nil
	}

	trace := tempBuf[:n]

	lineCount := 0
	lastNewline := -1
	for i := 0; i < len(trace); i++ {
		if trace[i] == '\n' {
			lineCount++
			lastNewline = i
			if lineCount > (2*depth + 1) {
				break
			}
		}
	}

	if lastNewline != -1 && lineCount > (2*depth+1) {
		trace = trace[:lastNewline]
	}

	return trace, bufPtr
}
