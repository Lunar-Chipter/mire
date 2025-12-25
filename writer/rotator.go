package writer

import (
	"os"
	"sync"

	"github.com/Lunar-Chipter/mire/config"
)

// Rotator provides a file writer that rotates logs.
type Rotator struct {
	file   *os.File
	closed bool
	mu     sync.Mutex
}

// NewRotator creates a new rotating file writer.
func NewRotator(filename string, conf *config.RotationConfig) (*Rotator, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Rotator{
		file:   f,
		closed: false,
	}, nil
}

// Write writes data to file, handling rotation if necessary.
func (w *Rotator) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)
	return n, err
}

// Close closes the underlying file.
func (w *Rotator) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	err := w.file.Close()
	w.closed = true
	return err
}
