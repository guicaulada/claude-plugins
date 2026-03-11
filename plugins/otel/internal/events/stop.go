package events

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	pluginotel "github.com/guicaulada/claude-plugins/plugins/otel/internal/otel"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/state"
)

func HandleStop(env payload.Envelope) error {
	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

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
	provider, err := newProvider(ctx, cfg)
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

	// Load recorded events for this prompt span (tool calls, agent starts)
	spanEvents := loadSpanEvents(store, env.SessionID, prompt.SpanID)
	debug.Log("stop: loaded %d events for prompt span %s", len(spanEvents), prompt.SpanID)

	// Export orphaned children before the prompt span itself
	exportOrphanedSubagents(store, builder, env.SessionID, sess.TraceID, prompt.SpanID, endTime)
	exportOrphanedTools(store, builder, env.SessionID, sess.TraceID, prompt.SpanID, endTime)

	builder.CreateSpan(promptCtx, "prompt", startTime, endTime, attrs, spanEvents...)

	durationMs := endTime.Sub(startTime).Milliseconds()

	// Record event on session span timeline
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    sess.SpanID,
		Name:      "prompt.stop",
		Timestamp: endTime.UnixNano(),
		Attrs:     marshalAttrs(map[string]string{"prompt.index": fmt.Sprintf("%d", prompt.PromptIndex), "duration_ms": fmt.Sprintf("%d", durationMs)}),
	})

	// Emit event
	logAttrs := commonLogAttrs(env)
	logAttrs = append(logAttrs,
		log.Int("claude_code.prompt.index", prompt.PromptIndex),
		log.Int64("claude_code.prompt.duration_ms", durationMs),
	)
	provider.EmitEvent("claude_code.prompt.stop", sess.TraceID, prompt.SpanID, logAttrs)

	// Emit metric
	metricAttrs := cwdMetricAttr(env.Cwd, cfg.IncludeHighCardinality)
	metricAttrs = append(metricAttrs, vcsMetricAttrs(env.Cwd, cfg.IncludeHighCardinality)...)
	provider.HistogramRecord(ctx, "claude_code.prompt.duration", endTime.Sub(startTime).Seconds(), metricAttrs...)

	debug.Log("stop: exported prompt span session=%s index=%d duration=%dms",
		env.SessionID, prompt.PromptIndex, durationMs)

	// Clean up the prompt from state
	return store.DeletePrompt(prompt.ID)
}
