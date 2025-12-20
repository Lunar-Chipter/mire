//go:build with_metrics
// +build with_metrics

package metric

import (
	"testing"

	"github.com/Lunar-Chipter/mire/core"
)

func TestDefaultMetricsCollector_WithMetrics(t *testing.T) {
	collector := NewDefaultMetricsCollector()
	collector.IncrementCounter(core.INFO, nil)
	if collector.GetCounter("log.info") != 1 {
		t.Errorf("expected counter to be 1, got %d", collector.GetCounter("log.info"))
	}
}
