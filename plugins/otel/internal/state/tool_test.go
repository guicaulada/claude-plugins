package state

import "testing"

func TestToolCRUD(t *testing.T) {
	store, err := Open("test-tool-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() { _ = store.Cleanup() }()

	// No tool initially
	tool, err := store.GetTool("tool-1")
	if err != nil {
		t.Fatalf("GetTool: %v", err)
	}
	if tool.ToolUseID != "" {
		t.Error("expected empty tool")
	}

	// Create
	if err := store.CreateTool(Tool{
		ToolUseID:    "tool-1",
		SessionID:    "s1",
		SpanID:       "span-t1",
		ParentSpanID: "span-p1",
		StartTime:    100,
		ToolName:     "Edit",
		FilePath:     "/tmp/test.go",
	}); err != nil {
		t.Fatalf("CreateTool: %v", err)
	}

	// Get
	tool, err = store.GetTool("tool-1")
	if err != nil {
		t.Fatalf("GetTool: %v", err)
	}
	if tool.ToolName != "Edit" {
		t.Errorf("ToolName = %q, want %q", tool.ToolName, "Edit")
	}
	if tool.FilePath != "/tmp/test.go" {
		t.Errorf("FilePath = %q, want %q", tool.FilePath, "/tmp/test.go")
	}
	if tool.ParentSpanID != "span-p1" {
		t.Errorf("ParentSpanID = %q, want %q", tool.ParentSpanID, "span-p1")
	}

	// Delete
	if err := store.DeleteTool("tool-1"); err != nil {
		t.Fatalf("DeleteTool: %v", err)
	}
	tool, _ = store.GetTool("tool-1")
	if tool.ToolUseID != "" {
		t.Error("expected empty tool after delete")
	}
}
