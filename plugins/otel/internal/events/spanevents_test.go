package events

import (
	"encoding/json"
	"testing"
)

func TestMarshalAttrs(t *testing.T) {
	attrs := map[string]string{
		"tool.name":   "Edit",
		"duration_ms": "100",
	}
	result := marshalAttrs(attrs)

	var parsed map[string]string
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("marshalAttrs produced invalid JSON: %v", err)
	}
	if parsed["tool.name"] != "Edit" {
		t.Errorf("tool.name = %q, want %q", parsed["tool.name"], "Edit")
	}
}

func TestMarshalAttrsSpecialChars(t *testing.T) {
	attrs := map[string]string{
		"message": `contains "quotes" and \backslashes`,
	}
	result := marshalAttrs(attrs)

	var parsed map[string]string
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("marshalAttrs should handle special chars: %v", err)
	}
	if parsed["message"] != `contains "quotes" and \backslashes` {
		t.Errorf("message = %q", parsed["message"])
	}
}

func TestCurrentTimestamp(t *testing.T) {
	ts := currentTimestamp()
	if ts <= 0 {
		t.Error("currentTimestamp should return positive value")
	}
}
