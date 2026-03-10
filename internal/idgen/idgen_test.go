package idgen

import (
	"testing"
)

func TestTraceID(t *testing.T) {
	id := TraceID()
	if len(id) != 32 {
		t.Errorf("TraceID length = %d, want 32", len(id))
	}

	// Should be unique
	id2 := TraceID()
	if id == id2 {
		t.Error("two TraceIDs should not be equal")
	}
}

func TestSpanID(t *testing.T) {
	id := SpanID()
	if len(id) != 16 {
		t.Errorf("SpanID length = %d, want 16", len(id))
	}

	id2 := SpanID()
	if id == id2 {
		t.Error("two SpanIDs should not be equal")
	}
}

func TestTraceIDHexChars(t *testing.T) {
	id := TraceID()
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("TraceID contains non-hex char: %c", c)
		}
	}
}
