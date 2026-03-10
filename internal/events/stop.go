package events

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	pluginotel "github.com/guicaulada/claude-code-otel-plugin/internal/otel"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

func HandleStop(env payload.Envelope) error {
	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	sess, err := store.GetSession(env.SessionID)
	if err != nil || sess.SessionID == "" {
		debug.Log("stop: no session state for %s", env.SessionID)
		return err
	}

	prompt, err := store.GetCurrentPrompt(env.SessionID)
	if err != nil || prompt.SessionID == "" {
		debug.Log("stop: no active prompt for %s", env.SessionID)
		return err
	}

	// Create and export the prompt span
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProviderFromState(ctx, cfg, store)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	parentCtx, err := pluginotel.ParentContext(sess.TraceID, sess.SpanID)
	if err != nil {
		return err
	}

	builder := pluginotel.NewSpanBuilder(provider.Tracer())
	startTime := time.Unix(0, prompt.StartTime)
	endTime := time.Now()

	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", env.SessionID),
		attribute.Int("claude_code.prompt.index", prompt.PromptIndex),
		attribute.String("claude_code.permission_mode", env.PermissionMode),
	}

	builder.CreateSpan(parentCtx, "prompt", startTime, endTime, attrs)

	debug.Log("stop: exported prompt span session=%s index=%d duration=%dms",
		env.SessionID, prompt.PromptIndex,
		endTime.Sub(startTime).Milliseconds())

	// Clean up the prompt from state
	return store.DeletePrompt(prompt.ID)
}
