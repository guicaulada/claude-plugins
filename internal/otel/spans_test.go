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
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithIDGenerator(newFixedIDGenerator()),
	)
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

func TestRootSpanHasNoParent(t *testing.T) {
	tracer, recorder := newTestTracer()
	builder := NewSpanBuilder(tracer)

	traceIDHex := "0123456789abcdef0123456789abcdef"
	spanIDHex := "fedcba9876543210"

	rootCtx, err := RootContext(traceIDHex, spanIDHex)
	if err != nil {
		t.Fatalf("RootContext: %v", err)
	}

	start := time.Now()
	end := start.Add(time.Minute)

	actualSpanID := builder.CreateSpan(rootCtx, "session", start, end, nil)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	s := spans[0]

	// Root span must have the correct trace ID
	if s.SpanContext().TraceID().String() != traceIDHex {
		t.Errorf("trace ID = %q, want %q", s.SpanContext().TraceID().String(), traceIDHex)
	}

	// Root span must have the predetermined span ID (so children link correctly)
	if actualSpanID != spanIDHex {
		t.Errorf("span ID = %q, want %q", actualSpanID, spanIDHex)
	}

	// Root span must NOT have a parent
	if s.Parent().IsValid() {
		t.Errorf("root span should have no parent, got parent span ID = %q", s.Parent().SpanID())
	}
}

func TestRootAndChildLinking(t *testing.T) {
	tracer, recorder := newTestTracer()
	builder := NewSpanBuilder(tracer)

	traceIDHex := "aabbccdd11223344aabbccdd11223344"
	sessionSpanIDHex := "1122334455667788"

	// Create child span first (like tool spans exported during session)
	childCtx, _ := ParentContext(traceIDHex, sessionSpanIDHex)
	builder.CreateSpan(childCtx, "tool:Edit", time.Now(), time.Now().Add(time.Second), nil)

	// Create root span last (deferred session span)
	rootCtx, _ := RootContext(traceIDHex, sessionSpanIDHex)
	builder.CreateSpan(rootCtx, "session", time.Now(), time.Now().Add(time.Minute), nil)

	spans := recorder.Ended()
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}

	child := spans[0]
	root := spans[1]

	// Both share the same trace ID
	if child.SpanContext().TraceID() != root.SpanContext().TraceID() {
		t.Error("child and root should share the same trace ID")
	}

	// Child's parent span ID matches root's span ID
	if child.Parent().SpanID().String() != root.SpanContext().SpanID().String() {
		t.Errorf("child parent = %q, root span = %q — should match",
			child.Parent().SpanID(), root.SpanContext().SpanID())
	}

	// Root has no parent
	if root.Parent().IsValid() {
		t.Errorf("root should have no parent, got %q", root.Parent().SpanID())
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
