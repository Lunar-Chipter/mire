package writer

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"mire/util"
)

// wrappedError wraps an error with a message
type wrappedError struct {
	msg   string
	cause error
}

func (e *wrappedError) Error() string {
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

func (e *wrappedError) Unwrap() error {
	return e.cause
}

// BufferedWriter is a high-performance buffered writer with batch processing
// BufferedWriter adalah writer yang di-buffer berkinerja tinggi dengan pemrosesan batch
type BufferedWriter struct {
	writer        io.Writer
	buffer        chan []byte
	bufferSize    int
	flushInterval time.Duration
	done          chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	droppedLogs   int64
	totalLogs     int64
	lastFlush     time.Time
	bufferPool    sync.Pool
	batchSize     int
	batchTimeout  time.Duration
	errorHandler  func(error)
}

// NewBufferedWriter creates a new BufferedWriter
func NewBufferedWriter(writer io.Writer, bufferSize int, flushInterval time.Duration, errorHandler func(error), batchSize int, batchTimeout time.Duration) *BufferedWriter {
	bw := &BufferedWriter{
		writer:        writer,
		buffer:        make(chan []byte, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
		lastFlush:     time.Now(),
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
		errorHandler:  errorHandler,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024)
			},
		},
	}

	bw.wg.Add(1)
	go bw.flushWorker()

	return bw
}

// Write writes data to the buffer. If the buffer is full, the log is dropped.
func (bw *BufferedWriter) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&bw.totalLogs, 1)

	// In order to not retain the original buffer `p`, we must copy it.
	buf := bw.bufferPool.Get().([]byte)
	if cap(buf) < len(p) {
		buf = make([]byte, len(p))
	}
	buf = buf[:len(p)]
	copy(buf, p)

	select {
	case bw.buffer <- buf:
		// The buffer was successfully queued.
		return len(p), nil
	default:
		// The buffer channel is full. Drop the log to prevent blocking.
		atomic.AddInt64(&bw.droppedLogs, 1)
		// We must return the buffer to the pool since it was not sent.
		bw.bufferPool.Put(buf[:0])
		return len(p), nil
	}
}

// flushWorker is the goroutine that flushes buffered logs
func (bw *BufferedWriter) flushWorker() {
	defer bw.wg.Done()

	ticker := time.NewTicker(bw.flushInterval)
	defer ticker.Stop()

	batch := make([][]byte, 0, bw.batchSize)

	// Only create a timer if a timeout is specified.
	var batchTimer *time.Timer
	var batchTimeoutChan <-chan time.Time
	if bw.batchTimeout > 0 {
		batchTimer = time.NewTimer(bw.batchTimeout)
		defer batchTimer.Stop()
		// Stop the timer immediately; it will be reset when a batch forms.
		if !batchTimer.Stop() {
			select {
			case <-batchTimer.C:
			default:
			}
		}
		batchTimeoutChan = batchTimer.C
	}

	for {
		select {
		case <-bw.done:
			ticker.Stop() // Stop the main flush ticker
			if batchTimer != nil {
				if !batchTimer.Stop() {
					select {
					case <-batchTimer.C: // Drain the channel if it has fired
					default:
					}
				}
			}

			// Drain remaining messages from the buffer channel completely.
			// The bw.buffer channel is now closed by Close() method.
			for data := range bw.buffer { // This loop will now terminate naturally
				batch = append(batch, data)
				if bw.batchSize > 0 && len(batch) >= bw.batchSize {
					bw.flushBatch(batch)
					batch = batch[:0]
				}
			}
			// Flush any final partial batch.
			if len(batch) > 0 {
				bw.flushBatch(batch)
			}
			return // EXIT the worker goroutine

		case <-ticker.C:
			if len(batch) > 0 {
				bw.flushBatch(batch)
				batch = batch[:0]
				// Stop the batch timer as the batch is now empty.
				if batchTimer != nil && !batchTimer.Stop() {
					select {
					case <-batchTimer.C:
					default:
					}
				}
			}

		case data := <-bw.buffer:
			if len(batch) == 0 && batchTimer != nil {
				// This is the first item in a new batch, start the timeout timer.
				batchTimer.Reset(bw.batchTimeout)
			}
			batch = append(batch, data)
			if bw.batchSize > 0 && len(batch) >= bw.batchSize {
				bw.flushBatch(batch)
				batch = batch[:0]
				// Stop the batch timer as the batch is now full and flushed.
				if batchTimer != nil && !batchTimer.Stop() {
					select {
					case <-batchTimer.C:
					default:
					}
				}
			}

		case <-batchTimeoutChan:
			if len(batch) > 0 {
				bw.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch flushes a batch of log entries
func (bw *BufferedWriter) flushBatch(batch [][]byte) {
	if len(batch) == 0 {
		return
	}

	bw.mu.Lock()
	defer bw.mu.Unlock()

	totalSize := 0
	for _, data := range batch {
		totalSize += len(data)
	}

	combined := make([]byte, 0, totalSize)
	for _, data := range batch {
		combined = append(combined, data...)
		bw.bufferPool.Put(data[:0])
	}

	if _, err := bw.writer.Write(combined); err != nil {
		if bw.errorHandler != nil {
			bw.errorHandler(&wrappedError{
				msg:   "buffered writer error",
				cause: err,
			})
		} else {
			// Write error message manually without fmt
			os.Stderr.Write([]byte("Error writing from buffered writer: "))
			os.Stderr.Write(util.S2b(err.Error()))
			os.Stderr.Write([]byte("\n"))
		}
	}
}

// Stats returns statistics about the buffered writer
func (bw *BufferedWriter) Stats() map[string]interface{} {
	return map[string]interface{}{
		"buffer_size":   bw.bufferSize,
		"current_queue": len(bw.buffer),
		"dropped_logs":  atomic.LoadInt64(&bw.droppedLogs),
		"total_logs":    atomic.LoadInt64(&bw.totalLogs),
		"last_flush":    bw.lastFlush,
	}
}

// Close closes the buffered writer, ensuring all logs are flushed.
func (bw *BufferedWriter) Close() error {
	close(bw.done)
	close(bw.buffer) // Close the input channel to flushWorker

	bw.wg.Wait()

	// Hanya tutup writer yang mendasarinya jika bukan os.Stdout atau os.Stderr
	// karena objek-objek ini tidak boleh ditutup oleh aplikasi.
	if bw.writer != os.Stdout && bw.writer != os.Stderr {
		if closer, ok := bw.writer.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}
