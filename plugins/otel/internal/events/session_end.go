package events

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	pluginotel "github.com/guicaulada/claude-plugins/plugins/otel/internal/otel"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/state"
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
		_ = store.Close()
		return err
	}

	debug.Log("session end: found state trace=%s", sess.TraceID)

	// Export the deferred session root span
	ctx := context.Background()
	cfg := config.Load()

	debug.Log("session end: creating OTel provider")
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		debug.Log("session end: failed to create provider: %v", err)
		_ = store.Close()
		return err
	}
	defer provider.Shutdown(ctx)

	startTime := time.Unix(0, sess.StartTime)
	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	builder := pluginotel.NewSpanBuilder(provider.Tracer())

	// Export orphaned prompts first (interrupted, never got Stop)
	// Must be before tools/subagents since they reference prompt as parent
	if orphanedPrompts, err := store.GetOrphanedPrompts(env.SessionID); err == nil && len(orphanedPrompts) > 0 {
		debug.Log("session end: exporting %d orphaned prompt spans", len(orphanedPrompts))
		for _, p := range orphanedPrompts {
			promptCtx, err := pluginotel.ChildContext(sess.TraceID, sess.SpanID, p.SpanID)
			if err != nil {
				continue
			}
			// Export orphaned children of this prompt
			exportOrphanedSubagents(store, builder, env.SessionID, sess.TraceID, p.SpanID, endTime)
			exportOrphanedTools(store, builder, env.SessionID, sess.TraceID, p.SpanID, endTime)

			promptEvents := loadSpanEvents(store, env.SessionID, p.SpanID)
			promptAttrs := []attribute.KeyValue{
				attribute.String("claude_code.session.id", env.SessionID),
				attribute.Int("claude_code.prompt.index", p.PromptIndex),
				attribute.Bool("claude_code.interrupted", true),
			}
			builder.CreateErrorSpan(promptCtx, "prompt", time.Unix(0, p.StartTime), endTime, promptAttrs, "interrupted", promptEvents...)
			_ = store.IncrementCounter(env.SessionID, "interrupted_count")
		}
	}

	// Export any remaining orphaned subagents (safety net for those not parented to a prompt)
	if orphanedAgents, err := store.GetOrphanedSubagents(env.SessionID); err == nil && len(orphanedAgents) > 0 {
		debug.Log("session end: exporting %d remaining orphaned subagent spans", len(orphanedAgents))
		for _, sa := range orphanedAgents {
			saCtx, err := pluginotel.ChildContext(sess.TraceID, sa.ParentSpanID, sa.SpanID)
			if err != nil {
				continue
			}
			exportOrphanedTools(store, builder, env.SessionID, sess.TraceID, sa.SpanID, endTime)

			saEvents := loadSpanEvents(store, env.SessionID, sa.SpanID)
			saAttrs := []attribute.KeyValue{
				attribute.String("claude_code.session.id", env.SessionID),
				attribute.String("claude_code.agent.type", sa.AgentType),
				attribute.String("claude_code.agent.name", sa.AgentName),
				attribute.String("claude_code.agent.id", sa.AgentID),
				attribute.Bool("claude_code.interrupted", true),
			}
			builder.CreateErrorSpan(saCtx, "agent:"+sa.AgentType, time.Unix(0, sa.StartTime), endTime, saAttrs, "interrupted", saEvents...)
			_ = store.IncrementCounter(env.SessionID, "interrupted_count")
		}
	}

	// Export any remaining orphaned tools (safety net for those not already exported)
	if orphanedTools, err := store.GetOrphanedTools(env.SessionID); err == nil && len(orphanedTools) > 0 {
		debug.Log("session end: exporting %d remaining orphaned tool spans", len(orphanedTools))
		for _, tool := range orphanedTools {
			toolCtx, err := pluginotel.ChildContext(sess.TraceID, tool.ParentSpanID, tool.SpanID)
			if err != nil {
				continue
			}
			toolEvents := loadSpanEvents(store, env.SessionID, tool.SpanID)
			toolAttrs := []attribute.KeyValue{
				attribute.String("claude_code.session.id", env.SessionID),
				attribute.String("claude_code.tool.name", tool.ToolName),
				attribute.String("claude_code.tool.use_id", tool.ToolUseID),
				attribute.Bool("claude_code.interrupted", true),
			}
			if tool.FilePath != "" {
				toolAttrs = append(toolAttrs, attribute.String("claude_code.file.path", tool.FilePath))
			}
			builder.CreateErrorSpan(toolCtx, "tool:"+tool.ToolName, time.Unix(0, tool.StartTime), endTime, toolAttrs, "interrupted", toolEvents...)
			_ = store.IncrementCounter(env.SessionID, "interrupted_count")
		}
	}

	// Read aggregated counters (after orphan export so interrupted_count is accurate)
	promptCount, _ := store.GetCounter(env.SessionID, "prompt_count")
	toolCount, _ := store.GetCounter(env.SessionID, "tool_count")
	errorCount, _ := store.GetCounter(env.SessionID, "error_count")
	subagentCount, _ := store.GetCounter(env.SessionID, "subagent_count")
	linesAdded, _ := store.GetCounter(env.SessionID, "lines_added")
	linesRemoved, _ := store.GetCounter(env.SessionID, "lines_removed")
	commitCount, _ := store.GetCounter(env.SessionID, "commit_count")
	branchCount, _ := store.GetCounter(env.SessionID, "branch_count")
	repoCount, _ := store.GetCounter(env.SessionID, "repo_count")
	interruptedCount, _ := store.GetCounter(env.SessionID, "interrupted_count")

	debug.Log("session end: counters prompts=%d tools=%d errors=%d subagents=%d lines_added=%d lines_removed=%d commits=%d branches=%d repos=%d interrupted=%d",
		promptCount, toolCount, errorCount, subagentCount, linesAdded, linesRemoved, commitCount, branchCount, repoCount, interruptedCount)

	debug.Log("session end: creating root context trace=%s span=%s", sess.TraceID, sess.SpanID)
	rootCtx, err := pluginotel.RootContext(sess.TraceID, sess.SpanID)
	if err != nil {
		debug.Log("session end: failed to create root context: %v", err)
		_ = store.Close()
		return err
	}

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
		attribute.Int64("claude_code.session.interrupted_count", interruptedCount),
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

	// Load recorded events for the session span
	spanEvents := loadSpanEvents(store, env.SessionID, sess.SpanID)
	debug.Log("session end: loaded %d span events", len(spanEvents))

	debug.Log("session end: creating session span")
	builder.CreateSpan(rootCtx, "session", startTime, endTime, attrs, spanEvents...)

	// Emit event
	logAttrs := commonLogAttrsFromSession(env, sess)
	logAttrs = append(logAttrs,
		log.String("claude_code.session.end_reason", event.Reason),
		log.String("claude_code.session.start_type", sess.StartType),
		log.Int64("claude_code.session.duration_ms", durationMs),
	)
	provider.EmitEvent("claude_code.session.end", sess.TraceID, sess.SpanID, logAttrs)

	// Emit metric
	metricAttrs := []attribute.KeyValue{
		attribute.String("claude_code.session.start_type", sess.StartType),
		attribute.String("claude_code.session.end_reason", event.Reason),
		attribute.String("claude_code.session.cwd", sess.Cwd),
	}
	metricAttrs = append(metricAttrs, vcsMetricAttrsFromSession(sess)...)
	provider.HistogramRecord(ctx, "claude_code.session.duration", endTime.Sub(startTime).Seconds(), metricAttrs...)

	debug.Log("session end: %s (trace: %s, duration: %dms, prompts: %d, tools: %d, errors: %d, subagents: %d lines_added=%d lines_removed=%d commits=%d branches=%d repos=%d)",
		env.SessionID, sess.TraceID, durationMs,
		promptCount, toolCount, errorCount, subagentCount,
		linesAdded, linesRemoved, commitCount, branchCount, repoCount)

	debug.Log("session end: cleaning up state")
	return store.Cleanup()
}
