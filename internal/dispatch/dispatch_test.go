package dispatch

import (
	"errors"
	"testing"

	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
)

func TestDispatchRegistered(t *testing.T) {
	r := New()
	called := false
	r.Register("SessionStart", func(env payload.Envelope) error {
		called = true
		if env.SessionID != "s1" {
			t.Errorf("SessionID = %q, want %q", env.SessionID, "s1")
		}
		return nil
	})

	env := payload.Envelope{
		HookEventName: "SessionStart",
		SessionID:     "s1",
	}

	if err := r.Dispatch(env); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDispatchUnregistered(t *testing.T) {
	r := New()

	env := payload.Envelope{
		HookEventName: "UnknownEvent",
		SessionID:     "s1",
	}

	if err := r.Dispatch(env); err != nil {
		t.Fatalf("Dispatch should not error for unknown events: %v", err)
	}
}

func TestDispatchError(t *testing.T) {
	r := New()
	r.Register("SessionEnd", func(env payload.Envelope) error {
		return errors.New("something went wrong")
	})

	env := payload.Envelope{
		HookEventName: "SessionEnd",
		SessionID:     "s1",
	}

	err := r.Dispatch(env)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "something went wrong" {
		t.Errorf("error = %q, want %q", err.Error(), "something went wrong")
	}
}
