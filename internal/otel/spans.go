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

// CreateSpan creates and immediately ends a span with explicit timestamps.
// Returns the span's hex span ID.
func (b *SpanBuilder) CreateSpan(ctx context.Context, name string, startTime, endTime time.Time, attrs []attribute.KeyValue) string {
	_, span := b.tracer.Start(ctx, name,
		trace.WithTimestamp(startTime),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
	span.End(trace.WithTimestamp(endTime))

	return span.SpanContext().SpanID().String()
}

// CreateErrorSpan creates and immediately ends an error span with explicit timestamps.
func (b *SpanBuilder) CreateErrorSpan(ctx context.Context, name string, startTime, endTime time.Time, attrs []attribute.KeyValue, errMsg string) string {
	_, span := b.tracer.Start(ctx, name,
		trace.WithTimestamp(startTime),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
	span.SetStatus(codes.Error, errMsg)
	span.End(trace.WithTimestamp(endTime))

	return span.SpanContext().SpanID().String()
}
