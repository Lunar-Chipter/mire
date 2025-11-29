// Package mire adalah package utama yang menyatukan semua modul
package mire

import (
	"github.com/Lunar-Chipter/mire/api"
	"github.com/Lunar-Chipter/mire/logger"
)

// New adalah fungsi utama untuk membuat instance logger baru
// Fungsi ini berfungsi sebagai titik masuk utama ke sistem logging Mire
func New(config logger.LoggerConfig) api.LoggerInterface {
	return logger.New(config)
}

// NewDefaultLogger adalah fungsi utama untuk membuat instance logger dengan konfigurasi default
func NewDefaultLogger() api.LoggerInterface {
	return logger.NewDefaultLogger()
}

// Module adalah interface untuk semua modul dalam Mire
type Module interface {
	Init() error
	Close() error
	Name() string
	Version() string
}

// CoreModule adalah modul untuk komponen inti
type CoreModule struct{}

func (c *CoreModule) Init() error { return nil }
func (c *CoreModule) Close() error { return nil }
func (c *CoreModule) Name() string { return "core" }
func (c *CoreModule) Version() string { return "1.0.0" }

// WriterModule adalah modul untuk komponen writer
type WriterModule struct{}

func (w *WriterModule) Init() error { return nil }
func (w *WriterModule) Close() error { return nil }
func (w *WriterModule) Name() string { return "writer" }
func (w *WriterModule) Version() string { return "1.0.0" }

// FormatterModule adalah modul untuk komponen formatter
type FormatterModule struct{}

func (f *FormatterModule) Init() error { return nil }
func (f *FormatterModule) Close() error { return nil }
func (f *FormatterModule) Name() string { return "formatter" }
func (f *FormatterModule) Version() string { return "1.0.0" }

// HookModule adalah modul untuk komponen hook
type HookModule struct{}

func (h *HookModule) Init() error { return nil }
func (h *HookModule) Close() error { return nil }
func (h *HookModule) Name() string { return "hook" }
func (h *HookModule) Version() string { return "1.0.0" }

// UtilModule adalah modul untuk komponen utilitas
type UtilModule struct{}

func (u *UtilModule) Init() error { return nil }
func (u *UtilModule) Close() error { return nil }
func (u *UtilModule) Name() string { return "util" }
func (u *UtilModule) Version() string { return "1.0.0" }

// GetModule mengembalikan modul berdasarkan nama
func GetModule(name string) Module {
	switch name {
	case "core":
		return &CoreModule{}
	case "writer":
		return &WriterModule{}
	case "formatter":
		return &FormatterModule{}
	case "hook":
		return &HookModule{}
	case "util":
		return &UtilModule{}
	default:
		return nil
	}
}