package state

import "testing"

func TestSubagentCRUD(t *testing.T) {
	store, err := Open("test-subagent-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() { _ = store.Cleanup() }()

	// No subagent initially
	sa, err := store.GetSubagent("agent-1")
	if err != nil {
		t.Fatalf("GetSubagent: %v", err)
	}
	if sa.AgentID != "" {
		t.Error("expected empty subagent")
	}

	// Create
	if err := store.CreateSubagent(Subagent{
		AgentID:      "agent-1",
		SessionID:    "s1",
		SpanID:       "span-a1",
		ParentSpanID: "span-p1",
		StartTime:    100,
		AgentType:    "Explore",
		AgentName:    "explorer",
	}); err != nil {
		t.Fatalf("CreateSubagent: %v", err)
	}

	// Get
	sa, err = store.GetSubagent("agent-1")
	if err != nil {
		t.Fatalf("GetSubagent: %v", err)
	}
	if sa.AgentType != "Explore" {
		t.Errorf("AgentType = %q, want %q", sa.AgentType, "Explore")
	}
	if sa.AgentName != "explorer" {
		t.Errorf("AgentName = %q, want %q", sa.AgentName, "explorer")
	}

	// Delete
	if err := store.DeleteSubagent("agent-1"); err != nil {
		t.Fatalf("DeleteSubagent: %v", err)
	}
	sa, _ = store.GetSubagent("agent-1")
	if sa.AgentID != "" {
		t.Error("expected empty subagent after delete")
	}
}
