package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
)

// CounterAdd increments a named counter metric with the given attributes.
func (p *Provider) CounterAdd(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	counter, err := p.Meter().Int64Counter(name)
	if err != nil {
		debug.Log("failed to create counter %s: %v", name, err)
		return
	}
	counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

// HistogramRecord records a value on a named histogram metric with the given attributes.
func (p *Provider) HistogramRecord(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	histogram, err := p.Meter().Float64Histogram(name)
	if err != nil {
		debug.Log("failed to create histogram %s: %v", name, err)
		return
	}
	histogram.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}
