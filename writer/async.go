package writer

import (
	"context"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"mire/core"
	"mire/errors"
	"mire/util"
)

// LogProcessor defines the interface for the underlying logger that the AsyncLogger will use.
// This helps to break the circular dependency between the writer and the main logger.
type LogProcessor interface {
	Log(ctx context.Context, level core.Level, msg []byte, fields map[string]interface{})
	ErrorHandler() func(error)
	ErrOut() io.Writer
	ErrOutMu() *sync.Mutex
}

// AsyncLogger provides asynchronous logging to reduce latency
type AsyncLogger struct {
	processor   LogProcessor
	logChan     chan *logJob
	wg          sync.WaitGroup
	workerCount int
	closed      atomic.Bool
	logProcessTimeout time.Duration
	disablePerLogContextTimeout bool
}

// logJob represents a logging job
type logJob struct {
	level  core.Level
	msg    []byte
	fields map[string]interface{}
	ctx    context.Context
}

// NewAsyncLogger creates a new AsyncLogger
func NewAsyncLogger(processor LogProcessor, workerCount int, bufferSize int, logProcessTimeout time.Duration, disablePerLogContextTimeout bool) *AsyncLogger {
	al := &AsyncLogger{
		processor:   processor,
		logChan:     make(chan *logJob, bufferSize),
		workerCount: workerCount,
		logProcessTimeout: logProcessTimeout,
		disablePerLogContextTimeout: disablePerLogContextTimeout,
	}

	for i := 0; i < workerCount; i++ {
		al.wg.Add(1)
		go al.worker()
	}

	return al
}

// worker is the goroutine that processes log jobs
func (al *AsyncLogger) worker() {
	defer al.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			if al.processor.ErrOut() != nil {
				// Use manual formatting to avoid fmt
				mu := al.processor.ErrOutMu()
				mu.Lock()
				defer mu.Unlock()

				// Write panic message manually
				al.processor.ErrOut().Write([]byte("recovering from panic in async logger worker: "))
				// Convert recovered value to string manually
				recoveredStr := util.ManualStringConversion(r)
				al.processor.ErrOut().Write(util.S2b(recoveredStr))
				al.processor.ErrOut().Write([]byte("\n"))

				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				al.processor.ErrOut().Write([]byte("stack trace: "))
				al.processor.ErrOut().Write(buf[:n])
				al.processor.ErrOut().Write([]byte("\n"))
			}
		}
	}()

	for job := range al.logChan {
		var ctx context.Context
		var cancel context.CancelFunc

		if al.logProcessTimeout > 0 && !al.disablePerLogContextTimeout {
			ctx, cancel = context.WithTimeout(job.ctx, al.logProcessTimeout)
		} else {
			ctx = job.ctx
		}

		al.processor.Log(ctx, job.level, job.msg, job.fields)

		if cancel != nil {
			cancel()
		}
	}
}

// Log queues a log job for asynchronous processing
func (al *AsyncLogger) Log(level core.Level, msg []byte, fields map[string]interface{}, ctx context.Context) {
	msgCopy := make([]byte, len(msg))
	copy(msgCopy, msg)

	select {
	case al.logChan <- &logJob{level: level, msg: msgCopy, fields: fields, ctx: ctx}:
		// Successfully sent
	default:
		// Channel full, handle error
		if handler := al.processor.ErrorHandler(); handler != nil {
			handler(errors.ErrAsyncBufferFull)
		} else if errOut := al.processor.ErrOut(); errOut != nil {
			mu := al.processor.ErrOutMu()
			mu.Lock()
			errOut.Write([]byte("Warning: Async log channel full, dropping log.\n"))
			mu.Unlock()
		}
	}
}

// Close closes the async logger
func (al *AsyncLogger) Close() {
	if al.closed.CompareAndSwap(false, true) {
		close(al.logChan)
		al.wg.Wait()
	}
}
