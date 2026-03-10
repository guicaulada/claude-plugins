package main

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/dispatch"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic recovered: %v", r)
		}
		os.Exit(0)
	}()

	cfg := config.Load()
	if !cfg.Enabled {
		return
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		debug.Log("failed to read stdin: %v", err)
		return
	}

	if len(input) == 0 {
		debug.Log("empty stdin, nothing to process")
		return
	}

	env, err := payload.Parse(input)
	if err != nil {
		debug.Log("failed to parse payload: %v", err)
		return
	}

	registry := dispatch.New()
	registerHandlers(registry)

	if err := registry.Dispatch(env); err != nil {
		debug.Log("handler error for %s: %v", env.HookEventName, err)
	}
}

func registerHandlers(r *dispatch.Registry) {
	r.Register("SessionStart", handleSessionStart)
	r.Register("SessionEnd", handleSessionEnd)
}

// sessionStartEvent holds SessionStart-specific fields.
type sessionStartEvent struct {
	Source string `json:"source"`
	Model  string `json:"model"`
}

func handleSessionStart(env payload.Envelope) error {
	var event sessionStartEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse SessionStart event: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	sess := state.Session{
		SessionID:      env.SessionID,
		TraceID:        generateID(32),
		SpanID:         generateID(16),
		StartTime:      time.Now().UnixNano(),
		Cwd:            env.Cwd,
		PermissionMode: env.PermissionMode,
		StartType:      event.Source,
	}

	if err := store.CreateSession(sess); err != nil {
		return err
	}

	debug.Log("session start: %s (trace: %s, type: %s, cwd: %s)",
		env.SessionID, sess.TraceID, event.Source, env.Cwd)
	return nil
}

func handleSessionEnd(env payload.Envelope) error {
	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}

	sess, err := store.GetSession(env.SessionID)
	if err != nil {
		store.Close()
		return err
	}

	if sess.SessionID == "" {
		debug.Log("session end: no state found for %s", env.SessionID)
		store.Close()
		return nil
	}

	debug.Log("session end: %s (trace: %s, duration: %dms)",
		env.SessionID, sess.TraceID,
		(time.Now().UnixNano()-sess.StartTime)/1e6)

	return store.Cleanup()
}

// generateID produces a hex string of the given length (16 for span_id, 32 for trace_id).
func generateID(hexLen int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, hexLen)
	// Use crypto-quality randomness would be ideal, but for span/trace IDs
	// we just need uniqueness. Read from urandom.
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// Fallback to time-based
		t := time.Now().UnixNano()
		for i := range b {
			b[i] = hex[(t>>uint(i*4))&0xf]
		}
		return string(b)
	}
	defer f.Close()

	raw := make([]byte, hexLen/2)
	f.Read(raw)
	for i, v := range raw {
		b[i*2] = hex[v>>4]
		b[i*2+1] = hex[v&0xf]
	}
	return string(b)
}
