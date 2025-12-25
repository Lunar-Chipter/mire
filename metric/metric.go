package metric

import (
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/Lunar-Chipter/mire/core"
)

// Collector interface defines methods for collecting metrics
type Collector interface {
	// IncrementCounter increments a counter metric
	IncrementCounter(level core.Level, tags map[string]string)

	// RecordHistogram records a histogram metric
	RecordHistogram(metric string, value float64, tags map[string]string)

	// RecordGauge records a gauge metric
	RecordGauge(metric string, value float64, tags map[string]string)
}

// Metrics is a simple in-memory metrics collector
type Metrics struct {
	counters   map[string]int64
	histograms map[string][]float64
	gauges     map[string]float64
	mu         sync.RWMutex
}

// NewMetrics creates a new Metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		counters:   make(map[string]int64),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

// IncrementCounter increments a counter metric
func (m *Metrics) IncrementCounter(level core.Level, tags map[string]string) {
	if level < core.TRACE || level > core.PANIC {
		return
	}
	key := "log." + strings.ToLower(level.String())
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[key]++
}

// RecordHistogram records a histogram metric
func (m *Metrics) RecordHistogram(metric string, value float64, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histograms[metric] = append(m.histograms[metric], value)
}

// RecordGauge records a gauge metric
func (m *Metrics) RecordGauge(metric string, value float64, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[metric] = value
}

// GetCounter returns the value of a counter metric
func (m *Metrics) GetCounter(metric string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[metric]
}

// GetHistogram returns statistics for a histogram metric
func (m *Metrics) GetHistogram(metric string) (min, max, avg, p95 float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := m.histograms[metric]
	if len(values) == 0 {
		return 0, 0, 0, 0
	}

	sort.Float64s(values)
	min = values[0]
	max = values[len(values)-1]

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg = sum / float64(len(values))

	p95Index := int(math.Ceil(0.95*float64(len(values)))) - 1
	if p95Index < 0 {
		p95Index = 0
	}
	p95 = values[p95Index]

	return min, max, avg, p95
}
