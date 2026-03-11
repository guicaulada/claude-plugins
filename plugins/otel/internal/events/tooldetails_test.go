package events

import (
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/log"
)

func findLogAttr(attrs []log.KeyValue, key string) (string, bool) {
	for _, kv := range attrs {
		if kv.Key == key {
			return kv.Value.AsString(), true
		}
	}
	return "", false
}

func TestAddToolInputDetails_Bash(t *testing.T) {
	var attrs []log.KeyValue
	input, _ := json.Marshal(map[string]any{
		"command":     "npm test",
		"description": "Run test suite",
	})
	addToolInputDetails(&attrs, "Bash", input)

	if v, ok := findLogAttr(attrs, "claude_code.tool.input.command"); !ok || v != "npm test" {
		t.Errorf("command = %q, want %q", v, "npm test")
	}
	if v, ok := findLogAttr(attrs, "claude_code.tool.input.description"); !ok || v != "Run test suite" {
		t.Errorf("description = %q, want %q", v, "Run test suite")
	}
}

func TestAddToolInputDetails_Edit(t *testing.T) {
	var attrs []log.KeyValue
	input, _ := json.Marshal(map[string]any{
		"file_path":  "/tmp/test.go",
		"old_string": "old",
		"new_string": "new",
	})
	addToolInputDetails(&attrs, "Edit", input)

	if v, ok := findLogAttr(attrs, "claude_code.tool.input.file_path"); !ok || v != "/tmp/test.go" {
		t.Errorf("file_path = %q, want %q", v, "/tmp/test.go")
	}
	// old_string and new_string should NOT be in attrs (sensitive content)
	if _, ok := findLogAttr(attrs, "claude_code.tool.input.old_string"); ok {
		t.Error("old_string should not be in attrs")
	}
}

func TestAddToolInputDetails_Grep(t *testing.T) {
	var attrs []log.KeyValue
	input, _ := json.Marshal(map[string]any{
		"pattern": "func Test",
		"path":    "/tmp",
		"glob":    "*.go",
	})
	addToolInputDetails(&attrs, "Grep", input)

	if v, _ := findLogAttr(attrs, "claude_code.tool.input.pattern"); v != "func Test" {
		t.Errorf("pattern = %q", v)
	}
	if v, _ := findLogAttr(attrs, "claude_code.tool.input.glob"); v != "*.go" {
		t.Errorf("glob = %q", v)
	}
}

func TestAddToolInputDetails_UnknownTool(t *testing.T) {
	var attrs []log.KeyValue
	addToolInputDetails(&attrs, "UnknownTool", []byte(`{"foo":"bar"}`))

	if len(attrs) != 0 {
		t.Errorf("expected no attrs for unknown tool, got %d", len(attrs))
	}
}

func TestAddToolInputDetails_EmptyInput(t *testing.T) {
	var attrs []log.KeyValue
	addToolInputDetails(&attrs, "Bash", nil)

	if len(attrs) != 0 {
		t.Errorf("expected no attrs for empty input, got %d", len(attrs))
	}
}
