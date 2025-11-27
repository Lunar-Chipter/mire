package metric

import (
	"math"
	"sort"
	"strings"
	"sync"

	"mire/core"
)

// MetricsCollector interface defines methods for collecting metrics
// Interface MetricsCollector mendefinisikan metode untuk mengumpulkan metrik
type MetricsCollector interface {
	// IncrementCounter increments a counter metric
	IncrementCounter(level core.Level, tags map[string]string)

	// RecordHistogram records a histogram metric
	RecordHistogram(metric string, value float64, tags map[string]string)

	// RecordGauge records a gauge metric
	RecordGauge(metric string, value float64, tags map[string]string)
}

// DefaultMetricsCollector is a simple in-memory metrics collector
// DefaultMetricsCollector adalah kolektor metrik dalam memori sederhana
type DefaultMetricsCollector struct {
	counters   map[string]int64
	histograms map[string][]float64
	gauges     map[string]float64
	mu         sync.RWMutex
}

// NewDefaultMetricsCollector creates a new DefaultMetricsCollector
func NewDefaultMetricsCollector() *DefaultMetricsCollector {
	return &DefaultMetricsCollector{
		counters:   make(map[string]int64),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

// IncrementCounter increments a counter metric
func (d *DefaultMetricsCollector) IncrementCounter(level core.Level, tags map[string]string) {
	if level < core.TRACE || level > core.PANIC {
		return // Invalid level
	}
	key := "log." + strings.ToLower(level.String())
	d.mu.Lock()
	defer d.mu.Unlock()
	d.counters[key]++
}

// RecordHistogram records a histogram metric
func (d *DefaultMetricsCollector) RecordHistogram(metric string, value float64, tags map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.histograms[metric] = append(d.histograms[metric], value)
}

// RecordGauge records a gauge metric
func (d *DefaultMetricsCollector) RecordGauge(metric string, value float64, tags map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gauges[metric] = value
}

// GetCounter returns the value of a counter metric
func (d *DefaultMetricsCollector) GetCounter(metric string) int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.counters[metric]
}

// GetHistogram returns statistics for a histogram metric
func (d *DefaultMetricsCollector) GetHistogram(metric string) (min, max, avg, p95 float64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	values := d.histograms[metric]
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
