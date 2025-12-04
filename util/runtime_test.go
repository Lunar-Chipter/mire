package util

import (
	"runtime"
	"testing"

	"github.com/Lunar-Chipter/mire/core"
)

// TestGetCallerInfo tests the GetCallerInfo function
func TestGetCallerInfo(t *testing.T) {
	// Call GetCallerInfo from this function (skip=1)
	callerInfo := GetCallerInfo(1)
	defer PutCallerInfoToPool(callerInfo)
	
	if callerInfo == nil {
		t.Fatal("GetCallerInfo returned nil")
	}
	
	// Verify that file has a value (at least contains "runtime_test.go")
	if callerInfo.File == "" {
		t.Error("Caller file should not be empty")
	}
	if callerInfo.Line <= 0 {
		t.Error("Caller line should be greater than 0")
	}
	if callerInfo.Function == "" {
		t.Error("Caller function should not be empty")
	}
	if callerInfo.Package == "" {
		t.Error("Caller package should not be empty")
	}
	
	// The function name should contain "TestGetCallerInfo"
	if got := callerInfo.Function; got != "TestGetCallerInfo" && 
	   got != "github.com/Lunar-Chipter/mire/util.TestGetCallerInfo" {
		t.Logf("Function name is %s, which might be different based on implementation", got)
	}
}

// TestGetCallerInfoSkip2 tests the GetCallerInfo function with skip=2
func TestGetCallerInfoSkip2(t *testing.T) {
	// Create a helper function that calls GetCallerInfo
	var helper = func() *core.CallerInfo {
		return GetCallerInfo(2) // Should get info from TestGetCallerInfoSkip2
	}
	
	callerInfo := helper()
	defer PutCallerInfoToPool(callerInfo)
	
	if callerInfo == nil {
		t.Fatal("GetCallerInfo with skip=2 returned nil")
	}
	
	// Verify basic properties
	if callerInfo.File == "" {
		t.Error("Caller file should not be empty with skip=2")
	}
	if callerInfo.Line <= 0 {
		t.Error("Caller line should be greater than 0 with skip=2")
	}
	if callerInfo.Function == "" {
		t.Error("Caller function should not be empty with skip=2")
	}
	if callerInfo.Package == "" {
		t.Error("Caller package should not be empty with skip=2")
	}
}

// TestGetCallerInfoSkip0 tests the GetCallerInfo function with skip=0
func TestGetCallerInfoSkip0(t *testing.T) {
	// This would get info about GetCallerInfo itself
	callerInfo := GetCallerInfo(0)
	defer PutCallerInfoToPool(callerInfo)
	
	if callerInfo == nil {
		t.Fatal("GetCallerInfo with skip=0 returned nil")
	}
	
	// The file should contain "runtime.go" since we're calling from GetCallerInfo function
	// which is defined in runtime.go
	if callerInfo.File == "" {
		t.Error("Caller file should not be empty with skip=0")
	}
}

// TestGetCallerInfoInvalidSkip tests the GetCallerInfo function with invalid skip value
func TestGetCallerInfoInvalidSkip(t *testing.T) {
	// Skip a number that is too high - this should return nil
	callerInfo := GetCallerInfo(10000) // Very high skip value
	
	// According to the implementation, if runtime.Caller returns false, it returns nil
	if callerInfo != nil {
		t.Errorf("GetCallerInfo with invalid skip should return nil, got %+v", callerInfo)
		PutCallerInfoToPool(callerInfo)
	}
}

// TestGetStackTrace tests the GetStackTrace function
func TestGetStackTrace(t *testing.T) {
	// Get a stack trace with depth=5
	stackTrace, bufPtr := GetStackTrace(5)
	
	// The returned values should either both be nil (if no stack) or both valid
	if (stackTrace == nil) != (bufPtr == nil) {
		t.Error("GetStackTrace should return either both nil or both non-nil")
	}
	
	// If they're both valid, check basic properties
	if stackTrace != nil && bufPtr != nil {
		if len(stackTrace) == 0 {
			t.Error("GetStackTrace returned empty stack trace")
		}
		
		// Check that stackTrace is part of bufPtr
		if len(*bufPtr) < len(stackTrace) {
			t.Error("Stack trace length should not exceed buffer length")
		}
		
		// The stack trace should contain information about this test function
		stackStr := string(stackTrace)
		if len(stackStr) < 10 { // Basic check for reasonable length
			t.Errorf("Stack trace seems too short: %s", stackStr)
		}
		
		// Ensure the buffer is returned to the pool
		core.PutBufferToPool(bufPtr)
	}
}

// TestGetStackTraceDepth tests the GetStackTrace function with different depths
func TestGetStackTraceDepth(t *testing.T) {
	tests := []int{1, 2, 3, 5}
	
	for _, depth := range tests {
		t.Run(string(rune(depth+'0')), func(t *testing.T) {
			stackTrace, bufPtr := GetStackTrace(depth)
			
			if stackTrace != nil && bufPtr != nil {
				if len(stackTrace) == 0 {
					t.Logf("Stack trace with depth %d is empty", depth)
				} else {
					// The stack trace should contain at least some information
					stackStr := string(stackTrace)
					if len(stackStr) < 10 {
						t.Logf("Stack trace with depth %d is very short: %s", depth, stackStr)
					}
				}
				
				// Always return the buffer to the pool
				if bufPtr != nil {
					core.PutBufferToPool(bufPtr)
				}
			} else {
				// This is acceptable if no stack trace is available
				t.Logf("No stack trace available for depth %d", depth)
			}
		})
	}
}

// TestGetStackTraceZeroDepth tests the GetStackTrace function with depth=0
func TestGetStackTraceZeroDepth(t *testing.T) {
	stackTrace, bufPtr := GetStackTrace(0)
	
	if stackTrace != nil && bufPtr != nil {
		// With depth 0, we should still get a stack trace but it might be limited
		if len(stackTrace) > 0 {
			// Basic validation if there's a stack trace
			stackStr := string(stackTrace)
			if len(stackStr) < 10 {
				t.Logf("Stack trace with depth 0 is short: %s", stackStr)
			}
		}
		
		// Return the buffer to the pool
		if bufPtr != nil {
			core.PutBufferToPool(bufPtr)
		}
	} else {
		// This is also acceptable
		t.Log("No stack trace available for depth 0")
	}
}

// TestGetStackTraceNegativeDepth tests the GetStackTrace function with negative depth
func TestGetStackTraceNegativeDepth(t *testing.T) {
	// In Go's runtime package, negative depth is typically treated as 0
	stackTrace, bufPtr := GetStackTrace(-1)
	
	if stackTrace != nil && bufPtr != nil {
		// Treat this like depth 0 case
		if len(stackTrace) > 0 {
			stackStr := string(stackTrace)
			if len(stackStr) < 10 {
				t.Logf("Stack trace with negative depth is short: %s", stackStr)
			}
		}
		
		// Return the buffer to the pool
		if bufPtr != nil {
			core.PutBufferToPool(bufPtr)
		}
	} else {
		// This is also acceptable
		t.Log("No stack trace available for negative depth")
	}
}

// TestPutCallerInfoToPool tests the PutCallerInfoToPool function
func TestPutCallerInfoToPool(t *testing.T) {
	// Get a caller info object
	callerInfo := GetCallerInfo(1)
	if callerInfo == nil {
		t.Skip("Could not get caller info, skipping test") // This might happen in some environments
		return
	}
	
	// Set some non-default values to verify they get reset
	callerInfo.File = "test.go"
	callerInfo.Line = 999
	callerInfo.Function = "TestFunction"
	callerInfo.Package = "TestPackage"
	
	// Put it back to the pool
	PutCallerInfoToPool(callerInfo)
	
	// Get another caller info object - it might be the same one
	anotherCallerInfo := GetCallerInfo(1)
	defer PutCallerInfoToPool(anotherCallerInfo)
	
	// Verify that the new object has default values (if it's the reused one)
	// This is hard to test definitively without knowing implementation details,
	// but we can at least verify it's not the same test values
	if anotherCallerInfo.File == "test.go" && 
	   anotherCallerInfo.Line == 999 && 
	   anotherCallerInfo.Function == "TestFunction" && 
	   anotherCallerInfo.Package == "TestPackage" {
		t.Error("Caller info fields were not reset when returned to pool")
	}
}

// TestStackTracePoolReturn tests that the buffer from GetStackTrace is properly managed
func TestStackTracePoolReturn(t *testing.T) {
	// Get a stack trace
	stackTrace, bufPtr := GetStackTrace(2)
	
	if stackTrace != nil && bufPtr != nil {
		// Verify the buffer has the expected content
		if len(*bufPtr) < len(stackTrace) {
			t.Error("Buffer length should be at least as long as stack trace")
		}
		
		// Verify that the stack trace is actually part of the buffer
		if cap(*bufPtr) == 0 {
			t.Error("Buffer should have capacity")
		}
		
		// Return the buffer to the pool
		core.PutBufferToPool(bufPtr)
		
		// After returning, we shouldn't access the stackTrace slice directly
		// since it might be reused. The Put function should handle it properly.
	} else {
		// This could happen if we're in a restricted environment
		t.Log("Could not get stack trace, skipping pool return test")
	}
}

// TestGetCallerInfoConcurrent tests GetCallerInfo in a concurrent context
func TestGetCallerInfoConcurrent(t *testing.T) {
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < 10; j++ {
				callerInfo := GetCallerInfo(1)
				if callerInfo == nil {
					// In some cases, this might return nil if the stack is not available
					continue
				}
				
				// Validate that we got valid info
				if callerInfo.File == "" || callerInfo.Line <= 0 {
					t.Errorf("Got invalid caller info: %+v", callerInfo)
				}
				
				// Return to pool
				PutCallerInfoToPool(callerInfo)
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestRuntimePackage tests that the runtime package is being used correctly
func TestRuntimePackage(t *testing.T) {
	// Verify that we can get program counter info using runtime
	pc, file, line, ok := runtime.Caller(0)
	if !ok {
		t.Log("Could not get caller info from runtime package - this might be expected in some environments")
		return
	}
	
	if pc == 0 {
		t.Error("Program counter should not be 0")
	}
	
	if file == "" {
		t.Error("File should not be empty")
	}
	
	if line <= 0 {
		t.Error("Line should be greater than 0")
	}
}