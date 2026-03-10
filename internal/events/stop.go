package events

import (
	"context"
	"fmt"
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

	// Use ChildContext so the prompt span gets the exact span ID from state
	// (which tool spans already reference as their parent)
	promptCtx, err := pluginotel.ChildContext(sess.TraceID, sess.SpanID, prompt.SpanID)
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

	// VCS enrichment (read fresh — branch/repo can change mid-session)
	attrs = append(attrs, vcsAttributes(env.Cwd, env.SessionID, store)...)

	builder.CreateSpan(promptCtx, "prompt", startTime, endTime, attrs)

	durationMs := endTime.Sub(startTime).Milliseconds()

	// Emit event
	provider.EmitEvent("claude_code.prompt.stop", sess.TraceID, prompt.SpanID, map[string]string{
		"claude_code.session.id":         env.SessionID,
		"claude_code.prompt.index":       fmt.Sprintf("%d", prompt.PromptIndex),
		"claude_code.prompt.duration_ms": fmt.Sprintf("%d", durationMs),
		"claude_code.permission_mode":    env.PermissionMode,
	})

	// Emit metric
	provider.HistogramRecord(ctx, "claude_code.prompt.duration", float64(durationMs),
		attribute.String("claude_code.session.cwd", env.Cwd),
	)

	debug.Log("stop: exported prompt span session=%s index=%d duration=%dms",
		env.SessionID, prompt.PromptIndex, durationMs)

	// Clean up the prompt from state
	return store.DeletePrompt(prompt.ID)
}
