package state

import "testing"

func TestPromptCRUD(t *testing.T) {
	store, err := Open("test-prompt-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() { _ = store.Cleanup() }()

	// No prompts initially
	p, err := store.GetCurrentPrompt("s1")
	if err != nil {
		t.Fatalf("GetCurrentPrompt: %v", err)
	}
	if p.SessionID != "" {
		t.Error("expected empty prompt")
	}

	// Create two prompts
	if err := store.CreatePrompt(Prompt{
		SessionID: "s1", SpanID: "span-1", StartTime: 100, PromptIndex: 1,
	}); err != nil {
		t.Fatalf("CreatePrompt 1: %v", err)
	}
	if err := store.CreatePrompt(Prompt{
		SessionID: "s1", SpanID: "span-2", StartTime: 200, PromptIndex: 2,
	}); err != nil {
		t.Fatalf("CreatePrompt 2: %v", err)
	}

	// GetCurrentPrompt returns the most recent
	p, err = store.GetCurrentPrompt("s1")
	if err != nil {
		t.Fatalf("GetCurrentPrompt: %v", err)
	}
	if p.SpanID != "span-2" {
		t.Errorf("SpanID = %q, want %q", p.SpanID, "span-2")
	}
	if p.PromptIndex != 2 {
		t.Errorf("PromptIndex = %d, want 2", p.PromptIndex)
	}

	// Delete
	if err := store.DeletePrompt(p.ID); err != nil {
		t.Fatalf("DeletePrompt: %v", err)
	}
	// After deleting the most recent, GetCurrentPrompt returns the older one
	p, err = store.GetCurrentPrompt("s1")
	if err != nil {
		t.Fatalf("GetCurrentPrompt after delete: %v", err)
	}
	if p.SpanID != "span-1" {
		t.Errorf("SpanID after delete = %q, want %q", p.SpanID, "span-1")
	}
}
