package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// SpanBuilder constructs and exports spans using explicit timestamps
// and cross-process parent context from stored state.
type SpanBuilder struct {
	tracer trace.Tracer
}

// NewSpanBuilder creates a SpanBuilder with the given tracer.
func NewSpanBuilder(tracer trace.Tracer) *SpanBuilder {
	return &SpanBuilder{tracer: tracer}
}

// ParentContext reconstructs a remote parent span context from stored IDs.
func ParentContext(traceIDHex, spanIDHex string) (context.Context, error) {
	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		return nil, err
	}
	spanID, err := trace.SpanIDFromHex(spanIDHex)
	if err != nil {
		return nil, err
	}

	parentCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	return trace.ContextWithRemoteSpanContext(context.Background(), parentCtx), nil
}

// RootContext creates a context for a root span that will have the given
// trace ID and no parent. The tracer generates a new span ID, so the
// returned context also carries the desired span ID for CreateRootSpan
// to use via the fixed ID generator.
func RootContext(traceIDHex, spanIDHex string) (context.Context, error) {
	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		return nil, err
	}
	spanID, err := trace.SpanIDFromHex(spanIDHex)
	if err != nil {
		return nil, err
	}

	// Store desired IDs in context for the fixed ID generator
	ctx := context.WithValue(context.Background(), fixedTraceIDKey{}, traceID)
	ctx = context.WithValue(ctx, fixedSpanIDKey{}, spanID)
	return ctx, nil
}

// ChildContext creates a context with a remote parent AND a predetermined
// span ID for the new span. Use this when the span ID was already generated
// (stored in state) and child spans reference it as their parent.
func ChildContext(traceIDHex, parentSpanIDHex, spanIDHex string) (context.Context, error) {
	ctx, err := ParentContext(traceIDHex, parentSpanIDHex)
	if err != nil {
		return nil, err
	}

	spanID, err := trace.SpanIDFromHex(spanIDHex)
	if err != nil {
		return nil, err
	}

	// Store the desired span ID for the fixed ID generator
	ctx = context.WithValue(ctx, fixedSpanIDKey{}, spanID)
	return ctx, nil
}

type fixedTraceIDKey struct{}
type fixedSpanIDKey struct{}

// SpanEvent represents a timestamped event to add to a span.
type SpanEvent struct {
	Name  string
	Time  time.Time
	Attrs []attribute.KeyValue
}

// CreateSpan creates and immediately ends a span with explicit timestamps.
// Optional events are added to the span before ending.
// Returns the span's hex span ID.
func (b *SpanBuilder) CreateSpan(ctx context.Context, name string, startTime, endTime time.Time, attrs []attribute.KeyValue, events ...SpanEvent) string {
	_, span := b.tracer.Start(ctx, name,
		trace.WithTimestamp(startTime),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)

	for _, e := range events {
		span.AddEvent(e.Name, trace.WithTimestamp(e.Time), trace.WithAttributes(e.Attrs...))
	}

	span.End(trace.WithTimestamp(endTime))

	return span.SpanContext().SpanID().String()
}

// CreateErrorSpan creates and immediately ends an error span with explicit timestamps.
func (b *SpanBuilder) CreateErrorSpan(ctx context.Context, name string, startTime, endTime time.Time, attrs []attribute.KeyValue, errMsg string, events ...SpanEvent) string {
	_, span := b.tracer.Start(ctx, name,
		trace.WithTimestamp(startTime),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)

	for _, e := range events {
		span.AddEvent(e.Name, trace.WithTimestamp(e.Time), trace.WithAttributes(e.Attrs...))
	}

	span.SetStatus(codes.Error, errMsg)
	span.End(trace.WithTimestamp(endTime))

	return span.SpanContext().SpanID().String()
}
