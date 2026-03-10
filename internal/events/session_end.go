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

func HandleSessionEnd(env payload.Envelope) error {
	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}

	sess, err := store.GetSession(env.SessionID)
	if err != nil || sess.SessionID == "" {
		debug.Log("session end: no state found for %s", env.SessionID)
		store.Close()
		return err
	}

	// Read aggregated counters
	promptCount, _ := store.GetCounter(env.SessionID, "prompt_count")
	toolCount, _ := store.GetCounter(env.SessionID, "tool_count")
	errorCount, _ := store.GetCounter(env.SessionID, "error_count")
	subagentCount, _ := store.GetCounter(env.SessionID, "subagent_count")

	// Export the deferred session root span
	ctx := context.Background()
	cfg := config.Load()
	provider, err := pluginotel.NewProvider(ctx, cfg)
	if err != nil {
		store.Close()
		return err
	}
	defer provider.Shutdown(ctx)

	startTime := time.Unix(0, sess.StartTime)
	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	// Session root span has no parent — use trace_id only
	rootCtx, err := pluginotel.RootContext(sess.TraceID, sess.SpanID)
	if err != nil {
		store.Close()
		return err
	}

	builder := pluginotel.NewSpanBuilder(provider.Tracer())
	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", env.SessionID),
		attribute.String("claude_code.session.start_type", sess.StartType),
		attribute.String("claude_code.session.cwd", sess.Cwd),
		attribute.String("claude_code.permission_mode", sess.PermissionMode),
		attribute.Int64("claude_code.session.prompt_count", promptCount),
		attribute.Int64("claude_code.session.tool_count", toolCount),
		attribute.Int64("claude_code.session.error_count", errorCount),
		attribute.Int64("claude_code.session.subagent_count", subagentCount),
		attribute.Int64("claude_code.session.duration_ms", durationMs),
	}

	if sess.GitBranch != "" {
		attrs = append(attrs, attribute.String("vcs.ref.head.name", sess.GitBranch))
	}
	if sess.GitRepoName != "" {
		attrs = append(attrs, attribute.String("vcs.repository.name", sess.GitRepoName))
	}
	if sess.GitRepoOwner != "" {
		attrs = append(attrs, attribute.String("vcs.repository.owner", sess.GitRepoOwner))
	}
	if sess.GitRemoteURL != "" {
		attrs = append(attrs, attribute.String("vcs.repository.url.full", sess.GitRemoteURL))
	}

	builder.CreateSpan(rootCtx, "session", startTime, endTime, attrs)

	debug.Log("session end: %s (trace: %s, duration: %dms, prompts: %d, tools: %d, errors: %d, subagents: %d)",
		env.SessionID, sess.TraceID, durationMs,
		promptCount, toolCount, errorCount, subagentCount)

	return store.Cleanup()
}
