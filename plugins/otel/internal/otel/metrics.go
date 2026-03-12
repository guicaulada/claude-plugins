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
	desc string
}

// histogramDef defines a histogram metric.
type histogramDef struct {
	name    string
	unit    string
	desc    string
	buckets []float64
}

// Metric name constants used in both registration and call sites.
const (
	MetricSessions           = "claude_code.sessions"
	MetricPrompts            = "claude_code.prompts"
	MetricToolUses           = "claude_code.tool.uses"
	MetricErrors             = "claude_code.errors"
	MetricLinesChanged       = "claude_code.lines_changed"
	MetricSubagents          = "claude_code.subagents"
	MetricCompacts           = "claude_code.compacts"
	MetricNotifications      = "claude_code.notifications"
	MetricTasks              = "claude_code.tasks"
	MetricPermissionRequests = "claude_code.permission_requests"

	MetricSessionDuration  = "claude_code.session.duration"
	MetricPromptDuration   = "claude_code.prompt.duration"
	MetricToolDuration     = "claude_code.tool.duration"
	MetricSubagentDuration = "claude_code.subagent.duration"
)

// Bucket boundaries in seconds, tailored to each metric's expected distribution.
var (
	// Tool durations: ~50ms to a few minutes
	toolBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300}
	// Session durations: seconds to hours
	sessionBuckets = []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600, 7200}
)

var counterDefs = []counterDef{
	{MetricSessions, "{session}", "Number of sessions started"},
	{MetricPrompts, "{prompt}", "Number of prompts submitted"},
	{MetricToolUses, "{use}", "Number of tool invocations"},
	{MetricErrors, "{error}", "Number of tool errors"},
	{MetricLinesChanged, "{line}", "Lines added or removed"},
	{MetricSubagents, "{agent}", "Number of subagents started"},
	{MetricCompacts, "{compact}", "Number of conversation compactions"},
	{MetricNotifications, "{notification}", "Number of notifications sent"},
	{MetricTasks, "{task}", "Number of tasks completed"},
	{MetricPermissionRequests, "{request}", "Number of permission requests"},
}

var histogramDefs = []histogramDef{
	{MetricSessionDuration, "s", "Session duration", sessionBuckets},
	{MetricPromptDuration, "s", "Time from prompt submit to response", sessionBuckets},
	{MetricToolDuration, "s", "Tool execution duration", toolBuckets},
	{MetricSubagentDuration, "s", "Subagent execution duration", toolBuckets},
}

func newInstruments(meter otelmetric.Meter) instruments {
	inst := instruments{
		counters:   make(map[string]otelmetric.Int64Counter, len(counterDefs)),
		histograms: make(map[string]otelmetric.Float64Histogram, len(histogramDefs)),
	}

	for _, d := range counterDefs {
		c, err := meter.Int64Counter(d.name,
			otelmetric.WithUnit(d.unit),
			otelmetric.WithDescription(d.desc),
		)
		if err != nil {
			debug.Log("failed to create counter %s: %v", d.name, err)
			continue
		}
		inst.counters[d.name] = c
	}

	for _, d := range histogramDefs {
		opts := []otelmetric.Float64HistogramOption{
			otelmetric.WithUnit(d.unit),
			otelmetric.WithDescription(d.desc),
		}
		if len(d.buckets) > 0 {
			opts = append(opts, otelmetric.WithExplicitBucketBoundaries(d.buckets...))
		}
		h, err := meter.Float64Histogram(d.name, opts...)
		if err != nil {
			debug.Log("failed to create histogram %s: %v", d.name, err)
			continue
		}
		inst.histograms[d.name] = h
	}

	return inst
}

// CounterAdd increments a named counter metric with the given attributes.
func (p *Provider) CounterAdd(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	counter, ok := p.instruments.counters[name]
	if !ok {
		debug.Log("counter %s not registered", name)
		return
	}
	counter.Add(ctx, value, otelmetric.WithAttributes(attrs...))
}

// HistogramRecord records a value on a named histogram metric with the given attributes.
func (p *Provider) HistogramRecord(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	histogram, ok := p.instruments.histograms[name]
	if !ok {
		debug.Log("histogram %s not registered", name)
		return
	}
	histogram.Record(ctx, value, otelmetric.WithAttributes(attrs...))
}
