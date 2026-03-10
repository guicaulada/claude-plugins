package otel

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newTestTracer() (trace.Tracer, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	return tp.Tracer("test"), recorder
}

func TestCreateSpan(t *testing.T) {
	tracer, recorder := newTestTracer()
	builder := NewSpanBuilder(tracer)

	start := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 10, 12, 0, 5, 0, time.UTC)

	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", "s1"),
		attribute.String("claude_code.tool.name", "Edit"),
	}

	spanID := builder.CreateSpan(context.Background(), "tool:Edit", start, end, attrs)

	if spanID == "" {
		t.Error("expected non-empty span ID")
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]
	if s.Name() != "tool:Edit" {
		t.Errorf("name = %q, want %q", s.Name(), "tool:Edit")
	}
	if s.StartTime() != start {
		t.Errorf("start = %v, want %v", s.StartTime(), start)
	}
	if s.EndTime() != end {
		t.Errorf("end = %v, want %v", s.EndTime(), end)
	}

	// Check attributes
	attrMap := make(map[string]string)
	for _, a := range s.Attributes() {
		attrMap[string(a.Key)] = a.Value.AsString()
	}
	if attrMap["claude_code.session.id"] != "s1" {
		t.Errorf("session.id = %q, want %q", attrMap["claude_code.session.id"], "s1")
	}
	if attrMap["claude_code.tool.name"] != "Edit" {
		t.Errorf("tool.name = %q, want %q", attrMap["claude_code.tool.name"], "Edit")
	}
}

func TestCreateSpanWithParent(t *testing.T) {
	tracer, recorder := newTestTracer()
	builder := NewSpanBuilder(tracer)

	parentCtx, err := ParentContext("0123456789abcdef0123456789abcdef", "0123456789abcdef")
	if err != nil {
		t.Fatalf("ParentContext: %v", err)
	}

	start := time.Now()
	end := start.Add(time.Second)

	builder.CreateSpan(parentCtx, "prompt", start, end, nil)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]
	// Verify parent trace ID matches
	if s.SpanContext().TraceID().String() != "0123456789abcdef0123456789abcdef" {
		t.Errorf("trace ID = %q, want %q", s.SpanContext().TraceID().String(), "0123456789abcdef0123456789abcdef")
	}
	// Verify parent span ID is set
	if s.Parent().SpanID().String() != "0123456789abcdef" {
		t.Errorf("parent span ID = %q, want %q", s.Parent().SpanID().String(), "0123456789abcdef")
	}
}

func TestCreateErrorSpan(t *testing.T) {
	tracer, recorder := newTestTracer()
	builder := NewSpanBuilder(tracer)

	start := time.Now()
	end := start.Add(time.Second)

	builder.CreateErrorSpan(context.Background(), "tool:Bash", start, end, nil, "command failed")

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]
	if s.Status().Code.String() != "Error" {
		t.Errorf("status = %q, want Error", s.Status().Code.String())
	}
	if s.Status().Description != "command failed" {
		t.Errorf("status desc = %q, want %q", s.Status().Description, "command failed")
	}
}

func TestParentContextInvalidIDs(t *testing.T) {
	_, err := ParentContext("invalid", "0123456789abcdef")
	if err == nil {
		t.Error("expected error for invalid trace ID")
	}

	_, err = ParentContext("0123456789abcdef0123456789abcdef", "invalid")
	if err == nil {
		t.Error("expected error for invalid span ID")
	}
}
