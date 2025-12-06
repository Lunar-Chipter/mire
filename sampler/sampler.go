package sampler

import (
	"context"
	"sync"
	"sync/atomic"
	"github.com/Lunar-Chipter/mire/core"
)

// LogSampler defines the interface for a logger that can be sampled.
type LogSampler interface {
	Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte)
}

// SamplingLogger provides log sampling to reduce volume
type SamplingLogger struct {
	processor LogSampler
	rate      int
	counter   int64
	mu        sync.Mutex
}

// NewSamplingLogger creates a new SamplingLogger
func NewSamplingLogger(processor LogSampler, rate int) *SamplingLogger {
	return &SamplingLogger{
		processor: processor,
		rate:      rate,
	}
}

// ShouldLog determines if a log should be recorded based on sampling rate
func (sl *SamplingLogger) ShouldLog() bool {
	if sl.rate <= 1 {
		return true
	}
	counter := atomic.AddInt64(&sl.counter, 1)
	return counter%int64(sl.rate) == 0
}

// Log logs a message if it passes the sampling rate.
func (sl *SamplingLogger) Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
    if sl.ShouldLog() {
        sl.processor.Log(ctx, level, msg, fields)
    }
}
