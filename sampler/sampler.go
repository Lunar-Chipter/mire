package sampler

import (
	"context"
	"github.com/Lunar-Chipter/mire/core"
	"sync/atomic"
)

// Sampler defines the interface for a logger that can be sampled.
type Sampler interface {
	Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte)
}

// LogSampler provides log sampling to reduce volume
type LogSampler struct {
	processor Sampler
	rate      int
	counter   int64
}

// NewSampler creates a new LogSampler
func NewSampler(processor Sampler, rate int) *LogSampler {
	return &LogSampler{
		processor: processor,
		rate:      rate,
	}
}

// ShouldLog determines if a log should be recorded based on sampling rate
func (ls *LogSampler) ShouldLog() bool {
	if ls.rate <= 1 {
		return true
	}
	counter := atomic.AddInt64(&ls.counter, 1)
	return counter%int64(ls.rate) == 0
}

// Log logs a message if it passes sampling rate.
func (ls *LogSampler) Log(ctx context.Context, level core.Level, msg []byte, fields map[string][]byte) {
	if ls.ShouldLog() {
		ls.processor.Log(ctx, level, msg, fields)
	}
}
