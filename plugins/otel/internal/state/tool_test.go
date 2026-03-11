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

func TestGetToolsByParent(t *testing.T) {
	store, err := Open("test-tools-parent-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() { _ = store.Cleanup() }()

	// No tools initially
	tools, err := store.GetToolsByParent("s1", "span-p1")
	if err != nil {
		t.Fatalf("GetToolsByParent: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}

	// Create tools with different parents
	for _, tool := range []Tool{
		{ToolUseID: "t1", SessionID: "s1", SpanID: "span-t1", ParentSpanID: "span-p1", StartTime: 200, ToolName: "Bash"},
		{ToolUseID: "t2", SessionID: "s1", SpanID: "span-t2", ParentSpanID: "span-p1", StartTime: 100, ToolName: "Read"},
		{ToolUseID: "t3", SessionID: "s1", SpanID: "span-t3", ParentSpanID: "span-p2", StartTime: 300, ToolName: "Edit"},
	} {
		if err := store.CreateTool(tool); err != nil {
			t.Fatalf("CreateTool(%s): %v", tool.ToolUseID, err)
		}
	}

	// Query by parent
	tools, err = store.GetToolsByParent("s1", "span-p1")
	if err != nil {
		t.Fatalf("GetToolsByParent: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	// Should be ordered by start_time ASC
	if tools[0].ToolUseID != "t2" {
		t.Errorf("first tool = %q, want t2 (earliest start_time)", tools[0].ToolUseID)
	}
	if tools[1].ToolUseID != "t1" {
		t.Errorf("second tool = %q, want t1", tools[1].ToolUseID)
	}

	// Different parent
	tools, err = store.GetToolsByParent("s1", "span-p2")
	if err != nil {
		t.Fatalf("GetToolsByParent: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].ToolUseID != "t3" {
		t.Errorf("tool = %q, want t3", tools[0].ToolUseID)
	}
}
