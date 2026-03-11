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

func TestGetSubagentsByParent(t *testing.T) {
	store, err := Open("test-subagents-parent-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() { _ = store.Cleanup() }()

	// No subagents initially
	agents, err := store.GetSubagentsByParent("s1", "span-p1")
	if err != nil {
		t.Fatalf("GetSubagentsByParent: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected 0 subagents, got %d", len(agents))
	}

	// Create subagents with different parents
	for _, sa := range []Subagent{
		{AgentID: "a1", SessionID: "s1", SpanID: "span-a1", ParentSpanID: "span-p1", StartTime: 200, AgentType: "Explore", AgentName: "explorer"},
		{AgentID: "a2", SessionID: "s1", SpanID: "span-a2", ParentSpanID: "span-p1", StartTime: 100, AgentType: "Plan", AgentName: "planner"},
		{AgentID: "a3", SessionID: "s1", SpanID: "span-a3", ParentSpanID: "span-p2", StartTime: 300, AgentType: "Explore", AgentName: "other"},
	} {
		if err := store.CreateSubagent(sa); err != nil {
			t.Fatalf("CreateSubagent(%s): %v", sa.AgentID, err)
		}
	}

	// Query by parent
	agents, err = store.GetSubagentsByParent("s1", "span-p1")
	if err != nil {
		t.Fatalf("GetSubagentsByParent: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 subagents, got %d", len(agents))
	}
	// Should be ordered by start_time ASC
	if agents[0].AgentID != "a2" {
		t.Errorf("first agent = %q, want a2 (earliest start_time)", agents[0].AgentID)
	}
	if agents[1].AgentID != "a1" {
		t.Errorf("second agent = %q, want a1", agents[1].AgentID)
	}

	// Different parent
	agents, err = store.GetSubagentsByParent("s1", "span-p2")
	if err != nil {
		t.Fatalf("GetSubagentsByParent: %v", err)
	}
	if len(agents) != 1 {
		t.Fatalf("expected 1 subagent, got %d", len(agents))
	}
	if agents[0].AgentID != "a3" {
		t.Errorf("agent = %q, want a3", agents[0].AgentID)
	}
}
