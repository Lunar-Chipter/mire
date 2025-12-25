package config

import (
	"time"
)

// RotationConfig holds configuration for log rotation
type RotationConfig struct {
	MaxSize         int64
	MaxAge          time.Duration
	MaxBackups      int
	LocalTime       bool
	Compress        bool
	RotateInterval  time.Duration
	FilenamePattern string
}
