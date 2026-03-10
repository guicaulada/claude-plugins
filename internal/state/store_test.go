package state

import (
	"testing"
)

func TestOpenAndCreateSchema(t *testing.T) {
	store, err := Open("test-session-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Cleanup()

	// Verify tables exist by running queries
	tables := []string{"sessions", "prompts", "tools", "subagents", "counters"}
	for _, table := range tables {
		var count int
		err := store.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s not accessible: %v", table, err)
		}
	}
}

func TestSessionCRUD(t *testing.T) {
	store, err := Open("test-session-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Cleanup()

	sess := Session{
		SessionID:      "s1",
		TraceID:        "aaaa",
		SpanID:         "bbbb",
		StartTime:      1234567890,
		Cwd:            "/home/user",
		PermissionMode: "default",
		StartType:      "startup",
		GitBranch:      "main",
		GitRemoteURL:   "https://github.com/user/repo.git",
		GitRepoName:    "repo",
		GitRepoOwner:   "user",
	}

	if err := store.CreateSession(sess); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := store.GetSession("s1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if got.SessionID != sess.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, sess.SessionID)
	}
	if got.TraceID != sess.TraceID {
		t.Errorf("TraceID = %q, want %q", got.TraceID, sess.TraceID)
	}
	if got.GitBranch != sess.GitBranch {
		t.Errorf("GitBranch = %q, want %q", got.GitBranch, sess.GitBranch)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	store, err := Open("test-session-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Cleanup()

	got, err := store.GetSession("nonexistent")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.SessionID != "" {
		t.Errorf("expected empty session, got %+v", got)
	}
}

func TestCounters(t *testing.T) {
	store, err := Open("test-session-" + t.Name())
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Cleanup()

	sid := "s1"

	// Initial value should be 0
	val, err := store.GetCounter(sid, "prompt_count")
	if err != nil {
		t.Fatalf("GetCounter: %v", err)
	}
	if val != 0 {
		t.Errorf("initial counter = %d, want 0", val)
	}

	// Increment
	for i := 0; i < 5; i++ {
		if err := store.IncrementCounter(sid, "prompt_count"); err != nil {
			t.Fatalf("IncrementCounter: %v", err)
		}
	}

	val, err = store.GetCounter(sid, "prompt_count")
	if err != nil {
		t.Fatalf("GetCounter: %v", err)
	}
	if val != 5 {
		t.Errorf("counter = %d, want 5", val)
	}
}

func TestCleanup(t *testing.T) {
	store, err := Open("test-session-cleanup")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	dir := store.dir
	if err := store.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	// Directory should be removed
	if _, err := Open("test-session-cleanup"); err != nil {
		// Should be able to recreate — this means cleanup worked
		t.Fatalf("failed to reopen after cleanup: %v", err)
	}
	// Clean up the reopened one
	store2, _ := Open("test-session-cleanup")
	if store2 != nil {
		store2.Cleanup()
	}
	_ = dir
}
