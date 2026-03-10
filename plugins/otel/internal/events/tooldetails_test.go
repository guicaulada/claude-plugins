package events

import (
	"encoding/json"
	"testing"
)

func TestAddToolInputDetails_Bash(t *testing.T) {
	attrs := make(map[string]string)
	input, _ := json.Marshal(map[string]any{
		"command":     "npm test",
		"description": "Run test suite",
	})
	addToolInputDetails(attrs, "Bash", input)

	if attrs["claude_code.tool.input.command"] != "npm test" {
		t.Errorf("command = %q, want %q", attrs["claude_code.tool.input.command"], "npm test")
	}
	if attrs["claude_code.tool.input.description"] != "Run test suite" {
		t.Errorf("description = %q, want %q", attrs["claude_code.tool.input.description"], "Run test suite")
	}
}

func TestAddToolInputDetails_Edit(t *testing.T) {
	attrs := make(map[string]string)
	input, _ := json.Marshal(map[string]any{
		"file_path":  "/tmp/test.go",
		"old_string": "old",
		"new_string": "new",
	})
	addToolInputDetails(attrs, "Edit", input)

	if attrs["claude_code.tool.input.file_path"] != "/tmp/test.go" {
		t.Errorf("file_path = %q, want %q", attrs["claude_code.tool.input.file_path"], "/tmp/test.go")
	}
	// old_string and new_string should NOT be in attrs (sensitive content)
	if _, ok := attrs["claude_code.tool.input.old_string"]; ok {
		t.Error("old_string should not be in attrs")
	}
}

func TestAddToolInputDetails_Grep(t *testing.T) {
	attrs := make(map[string]string)
	input, _ := json.Marshal(map[string]any{
		"pattern": "func Test",
		"path":    "/tmp",
		"glob":    "*.go",
	})
	addToolInputDetails(attrs, "Grep", input)

	if attrs["claude_code.tool.input.pattern"] != "func Test" {
		t.Errorf("pattern = %q", attrs["claude_code.tool.input.pattern"])
	}
	if attrs["claude_code.tool.input.glob"] != "*.go" {
		t.Errorf("glob = %q", attrs["claude_code.tool.input.glob"])
	}
}

func TestAddToolInputDetails_UnknownTool(t *testing.T) {
	attrs := make(map[string]string)
	addToolInputDetails(attrs, "UnknownTool", []byte(`{"foo":"bar"}`))

	if len(attrs) != 0 {
		t.Errorf("expected no attrs for unknown tool, got %d", len(attrs))
	}
}

func TestAddToolInputDetails_EmptyInput(t *testing.T) {
	attrs := make(map[string]string)
	addToolInputDetails(attrs, "Bash", nil)

	if len(attrs) != 0 {
		t.Errorf("expected no attrs for empty input, got %d", len(attrs))
	}
}
