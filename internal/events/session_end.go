package events

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	pluginotel "github.com/guicaulada/claude-code-otel-plugin/internal/otel"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

type sessionEndEvent struct {
	Reason string `json:"reason"`
}

func HandleSessionEnd(env payload.Envelope) error {
	var event sessionEndEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse SessionEnd event: %v", err)
	}

	debug.Log("session end: opening state for %s", env.SessionID)

	store, err := state.Open(env.SessionID)
	if err != nil {
		debug.Log("session end: failed to open state: %v", err)
		return err
	}

	sess, err := store.GetSession(env.SessionID)
	if err != nil || sess.SessionID == "" {
		debug.Log("session end: no state found for %s (err: %v)", env.SessionID, err)
		store.Close()
		return err
	}

	debug.Log("session end: found state trace=%s", sess.TraceID)

	// Read aggregated counters
	promptCount, _ := store.GetCounter(env.SessionID, "prompt_count")
	toolCount, _ := store.GetCounter(env.SessionID, "tool_count")
	errorCount, _ := store.GetCounter(env.SessionID, "error_count")
	subagentCount, _ := store.GetCounter(env.SessionID, "subagent_count")
	linesAdded, _ := store.GetCounter(env.SessionID, "lines_added")
	linesRemoved, _ := store.GetCounter(env.SessionID, "lines_removed")
	commitCount, _ := store.GetCounter(env.SessionID, "commit_count")
	branchCount, _ := store.GetCounter(env.SessionID, "branch_count")
	repoCount, _ := store.GetCounter(env.SessionID, "repo_count")

	debug.Log("session end: counters prompts=%d tools=%d errors=%d subagents=%d lines_added=%d lines_removed=%d commits=%d branches=%d repos=%d",
		promptCount, toolCount, errorCount, subagentCount, linesAdded, linesRemoved, commitCount, branchCount, repoCount)

	// Export the deferred session root span
	ctx := context.Background()
	cfg := config.Load()

	debug.Log("session end: creating OTel provider")
	provider, err := newProviderFromState(ctx, cfg, store)
	if err != nil {
		debug.Log("session end: failed to create provider: %v", err)
		store.Close()
		return err
	}
	defer provider.Shutdown(ctx)

	startTime := time.Unix(0, sess.StartTime)
	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	debug.Log("session end: creating root context trace=%s span=%s", sess.TraceID, sess.SpanID)
	rootCtx, err := pluginotel.RootContext(sess.TraceID, sess.SpanID)
	if err != nil {
		debug.Log("session end: failed to create root context: %v", err)
		store.Close()
		return err
	}

	builder := pluginotel.NewSpanBuilder(provider.Tracer())
	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", env.SessionID),
		attribute.String("claude_code.session.start_type", sess.StartType),
		attribute.String("claude_code.session.cwd", sess.Cwd),
		attribute.String("claude_code.session.end_reason", event.Reason),
		attribute.Int64("claude_code.session.prompt_count", promptCount),
		attribute.Int64("claude_code.session.tool_count", toolCount),
		attribute.Int64("claude_code.session.error_count", errorCount),
		attribute.Int64("claude_code.session.subagent_count", subagentCount),
		attribute.Int64("claude_code.session.lines_added", linesAdded),
		attribute.Int64("claude_code.session.lines_removed", linesRemoved),
		attribute.Int64("claude_code.session.commit_count", commitCount),
		attribute.Int64("claude_code.session.branch_count", branchCount),
		attribute.Int64("claude_code.session.repo_count", repoCount),
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

	debug.Log("session end: creating session span")
	builder.CreateSpan(rootCtx, "session", startTime, endTime, attrs)

	// Emit event
	provider.EmitEvent("claude_code.session.end", sess.TraceID, sess.SpanID, map[string]string{
		"claude_code.session.id":         env.SessionID,
		"claude_code.session.end_reason": event.Reason,
		"claude_code.session.cwd":        sess.Cwd,
	})

	// Emit metric
	provider.HistogramRecord(ctx, "claude_code.session.duration", float64(durationMs),
		attribute.String("claude_code.session.start_type", sess.StartType),
		attribute.String("claude_code.session.cwd", sess.Cwd),
	)

	debug.Log("session end: %s (trace: %s, duration: %dms, prompts: %d, tools: %d, errors: %d, subagents: %d lines_added=%d lines_removed=%d commits=%d branches=%d repos=%d)",
		env.SessionID, sess.TraceID, durationMs,
		promptCount, toolCount, errorCount, subagentCount,
		linesAdded, linesRemoved, commitCount, branchCount, repoCount)

	debug.Log("session end: cleaning up state")
	return store.Cleanup()
}
