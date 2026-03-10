package events

import (
	"time"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

func HandleSessionEnd(env payload.Envelope) error {
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
