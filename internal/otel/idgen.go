package otel

import (
	"context"
	"crypto/rand"

	"go.opentelemetry.io/otel/trace"
)

// fixedIDGenerator returns predetermined trace and span IDs from context.
// Falls back to random generation for IDs not in context.
type fixedIDGenerator struct{}

func newFixedIDGenerator() *fixedIDGenerator {
	return &fixedIDGenerator{}
}

func (g *fixedIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	traceID, ok := ctx.Value(fixedTraceIDKey{}).(trace.TraceID)
	if !ok {
		traceID = randomTraceID()
	}

	spanID, ok := ctx.Value(fixedSpanIDKey{}).(trace.SpanID)
	if !ok {
		spanID = randomSpanID()
	}

	return traceID, spanID
}

func (g *fixedIDGenerator) NewSpanID(ctx context.Context, _ trace.TraceID) trace.SpanID {
	spanID, ok := ctx.Value(fixedSpanIDKey{}).(trace.SpanID)
	if !ok {
		return randomSpanID()
	}
	return spanID
}

func randomTraceID() trace.TraceID {
	var id trace.TraceID
	_, _ = rand.Read(id[:])
	return id
}

func randomSpanID() trace.SpanID {
	var id trace.SpanID
	_, _ = rand.Read(id[:])
	return id
}
