package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
)

// CounterAdd increments a named counter metric with the given attributes.
// Cardinality-controlled base attributes (app.version) are automatically appended.
func (p *Provider) CounterAdd(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	counter, err := p.Meter().Int64Counter(name)
	if err != nil {
		debug.Log("failed to create counter %s: %v", name, err)
		return
	}
	allAttrs := append(attrs, p.metricBaseAttrs...)
	counter.Add(ctx, value, otelmetric.WithAttributes(allAttrs...))
}

// HistogramRecord records a value on a named histogram metric with the given attributes.
// Cardinality-controlled base attributes (app.version) are automatically appended.
func (p *Provider) HistogramRecord(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	histogram, err := p.Meter().Float64Histogram(name)
	if err != nil {
		debug.Log("failed to create histogram %s: %v", name, err)
		return
	}
	allAttrs := append(attrs, p.metricBaseAttrs...)
	histogram.Record(ctx, value, otelmetric.WithAttributes(allAttrs...))
}
