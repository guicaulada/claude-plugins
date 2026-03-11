package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
)

// instruments holds pre-registered metric instruments.
type instruments struct {
	counters   map[string]otelmetric.Int64Counter
	histograms map[string]otelmetric.Float64Histogram
}

// counterDef defines a counter metric.
type counterDef struct {
	name string
	unit string
}

// histogramDef defines a histogram metric.
type histogramDef struct {
	name string
	unit string
}

var counterDefs = []counterDef{
	{"claude_code.sessions", "{session}"},
	{"claude_code.prompts", "{prompt}"},
	{"claude_code.tool.uses", "{use}"},
	{"claude_code.errors", "{error}"},
	{"claude_code.lines_changed", "{line}"},
	{"claude_code.subagents", "{agent}"},
	{"claude_code.compacts", "{compact}"},
	{"claude_code.notifications", "{notification}"},
	{"claude_code.tasks", "{task}"},
}

var histogramDefs = []histogramDef{
	{"claude_code.session.duration", "s"},
	{"claude_code.prompt.duration", "s"},
	{"claude_code.tool.duration", "s"},
	{"claude_code.subagent.duration", "s"},
}

func newInstruments(meter otelmetric.Meter) instruments {
	inst := instruments{
		counters:   make(map[string]otelmetric.Int64Counter, len(counterDefs)),
		histograms: make(map[string]otelmetric.Float64Histogram, len(histogramDefs)),
	}

	for _, d := range counterDefs {
		c, err := meter.Int64Counter(d.name, otelmetric.WithUnit(d.unit))
		if err != nil {
			debug.Log("failed to create counter %s: %v", d.name, err)
			continue
		}
		inst.counters[d.name] = c
	}

	for _, d := range histogramDefs {
		h, err := meter.Float64Histogram(d.name, otelmetric.WithUnit(d.unit))
		if err != nil {
			debug.Log("failed to create histogram %s: %v", d.name, err)
			continue
		}
		inst.histograms[d.name] = h
	}

	return inst
}

// CounterAdd increments a named counter metric with the given attributes.
// Cardinality-controlled base attributes (app.version) are automatically appended.
func (p *Provider) CounterAdd(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	counter, ok := p.instruments.counters[name]
	if !ok {
		debug.Log("counter %s not registered", name)
		return
	}
	allAttrs := append(attrs, p.metricBaseAttrs...)
	counter.Add(ctx, value, otelmetric.WithAttributes(allAttrs...))
}

// HistogramRecord records a value on a named histogram metric with the given attributes.
// Cardinality-controlled base attributes (app.version) are automatically appended.
func (p *Provider) HistogramRecord(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	histogram, ok := p.instruments.histograms[name]
	if !ok {
		debug.Log("histogram %s not registered", name)
		return
	}
	allAttrs := append(attrs, p.metricBaseAttrs...)
	histogram.Record(ctx, value, otelmetric.WithAttributes(allAttrs...))
}
