package config

import (
	"time"
)

// RotationConfig holds configuration for log rotation
// RotationConfig menyimpan konfigurasi untuk rotasi log
type RotationConfig struct {
	MaxSize         int64
	MaxAge          time.Duration
	MaxBackups      int
	LocalTime       bool
	Compress        bool
	RotationTime    time.Duration
	FilenamePattern string
}
