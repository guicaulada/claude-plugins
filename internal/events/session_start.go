package events

import (
	"encoding/json"
	"time"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/idgen"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

type sessionStartEvent struct {
	Source string `json:"source"`
	Model  string `json:"model"`
}

func HandleSessionStart(env payload.Envelope) error {
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
		TraceID:        idgen.TraceID(),
		SpanID:         idgen.SpanID(),
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
