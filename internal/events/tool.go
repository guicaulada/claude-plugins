package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/fileinfo"
	gitpkg "github.com/guicaulada/claude-code-otel-plugin/internal/git"
	"github.com/guicaulada/claude-code-otel-plugin/internal/idgen"
	pluginotel "github.com/guicaulada/claude-code-otel-plugin/internal/otel"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

type toolEvent struct {
	ToolName  string          `json:"tool_name"`
	ToolUseID string          `json:"tool_use_id"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// editInput captures Edit tool fields for line diff computation.
type editInput struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

// writeInput captures Write tool fields.
type writeInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

func HandlePreToolUse(env payload.Envelope) error {
	var event toolEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse PreToolUse event: %v", err)
		return nil
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	// Determine parent span: subagent if inside one, otherwise current prompt
	parentSpanID := ""
	if env.AgentID != "" {
		sa, err := store.GetSubagent(env.AgentID)
		if err == nil && sa.AgentID != "" {
			parentSpanID = sa.SpanID
		}
	}
	if parentSpanID == "" {
		prompt, err := store.GetCurrentPrompt(env.SessionID)
		if err == nil && prompt.SessionID != "" {
			parentSpanID = prompt.SpanID
		}
	}
	if parentSpanID == "" {
		sess, err := store.GetSession(env.SessionID)
		if err == nil && sess.SessionID != "" {
			parentSpanID = sess.SpanID
		}
	}

	tool := state.Tool{
		ToolUseID:    event.ToolUseID,
		SessionID:    env.SessionID,
		SpanID:       idgen.SpanID(),
		ParentSpanID: parentSpanID,
		StartTime:    time.Now().UnixNano(),
		ToolName:     event.ToolName,
	}

	// Extract file_path from tool_input for file-based tools
	var genericInput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(event.ToolInput, &genericInput); err == nil && genericInput.FilePath != "" {
		tool.FilePath = genericInput.FilePath
	}

	// Snapshot file content for Write tool (to compute diffs in PostToolUse)
	if event.ToolName == "Write" && tool.FilePath != "" {
		if content, err := os.ReadFile(tool.FilePath); err == nil {
			tool.FileSnapshot = content
		}
	}

	// Cache HEAD SHA before Bash tool for commit detection
	if event.ToolName == "Bash" {
		sha := gitpkg.GetContext(env.Cwd).HeadSHA
		if sha != "" {
			_ = store.SetCache("git_head_sha_pre_"+event.ToolUseID, sha)
		}
	}

	if err := store.CreateTool(tool); err != nil {
		return err
	}

	// Emit event
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProviderFromState(ctx, cfg, store)
	if err == nil {
		defer provider.Shutdown(ctx)

		sess, _ := store.GetSession(env.SessionID)
		startLogAttrs := commonLogAttrs(env)
		startLogAttrs["claude_code.tool.name"] = event.ToolName
		startLogAttrs["claude_code.tool.use_id"] = event.ToolUseID
		if tool.FilePath != "" {
			startFi := fileinfo.FromPath(tool.FilePath)
			startLogAttrs["claude_code.file.path"] = startFi.Path
			startLogAttrs["claude_code.file.extension"] = startFi.Extension
			if startFi.Language != "" {
				startLogAttrs["claude_code.file.language"] = startFi.Language
			}
		}
		provider.EmitEvent("claude_code.tool.start", sess.TraceID, tool.SpanID, startLogAttrs)
	}

	// Record tool.start event on parent span
	if err := store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    parentSpanID,
		Name:      "tool.start",
		Timestamp: tool.StartTime,
		Attrs:     fmt.Sprintf(`{"tool.name":"%s"}`, event.ToolName),
	}); err != nil {
		debug.Log("failed to record tool.start event: %v", err)
	}

	debug.Log("pre tool use: session=%s tool=%s id=%s parent=%s",
		env.SessionID, event.ToolName, event.ToolUseID, parentSpanID)
	return nil
}

func HandlePostToolUse(env payload.Envelope) error {
	return handleToolEnd(env, false, "", false)
}

func HandlePostToolUseFailure(env payload.Envelope) error {
	var event struct {
		Error       string `json:"error"`
		IsInterrupt bool   `json:"is_interrupt"`
	}
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse PostToolUseFailure event: %v", err)
	}
	return handleToolEnd(env, true, event.Error, event.IsInterrupt)
}

func handleToolEnd(env payload.Envelope, isError bool, errMsg string, isInterrupt bool) error {
	var event toolEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse tool end event: %v", err)
		return nil
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	tool, err := store.GetTool(event.ToolUseID)
	if err != nil || tool.ToolUseID == "" {
		debug.Log("tool end: no state for tool_use_id=%s", event.ToolUseID)
		return err
	}

	sess, err := store.GetSession(env.SessionID)
	if err != nil || sess.SessionID == "" {
		debug.Log("tool end: no session state for %s", env.SessionID)
		return err
	}

	// Export the tool span
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProviderFromState(ctx, cfg, store)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	toolCtx, err := pluginotel.ChildContext(sess.TraceID, tool.ParentSpanID, tool.SpanID)
	if err != nil {
		return err
	}

	builder := pluginotel.NewSpanBuilder(provider.Tracer())
	startTime := time.Unix(0, tool.StartTime)
	endTime := time.Now()

	spanName := "tool:" + event.ToolName
	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", env.SessionID),
		attribute.String("claude_code.tool.name", event.ToolName),
		attribute.String("claude_code.tool.use_id", event.ToolUseID),
		attribute.String("claude_code.permission_mode", env.PermissionMode),
	}

	// VCS enrichment (read fresh — branch/repo can change mid-session)
	attrs = append(attrs, vcsAttributes(env.Cwd, env.SessionID, store)...)

	// File enrichment
	if tool.FilePath != "" {
		fi := fileinfo.FromPath(tool.FilePath)
		attrs = append(attrs,
			attribute.String("claude_code.file.path", fi.Path),
			attribute.String("claude_code.file.extension", fi.Extension),
		)
		if fi.Language != "" {
			attrs = append(attrs, attribute.String("claude_code.file.language", fi.Language))
		}
	}

	// Line diff computation for Edit and Write tools
	linesAdded, linesRemoved := computeLineDiff(event, tool, store)
	if linesAdded > 0 || linesRemoved > 0 {
		attrs = append(attrs,
			attribute.Int("claude_code.lines_added", linesAdded),
			attribute.Int("claude_code.lines_removed", linesRemoved),
		)
		_ = store.IncrementCounterBy(env.SessionID, "lines_added", int64(linesAdded))
		_ = store.IncrementCounterBy(env.SessionID, "lines_removed", int64(linesRemoved))
	}

	// Bash commit detection
	if event.ToolName == "Bash" {
		detectCommit(env, tool, store, &attrs)
	}

	// Load recorded events for this tool span (e.g., agent.start/stop for Agent tools)
	var spanEvents []pluginotel.SpanEvent
	if recorded, err := store.GetEvents(env.SessionID, tool.SpanID); err == nil && len(recorded) > 0 {
		debug.Log("tool end: loaded %d events for tool span %s", len(recorded), tool.SpanID)
		for _, re := range recorded {
			se := pluginotel.SpanEvent{
				Name: re.Name,
				Time: time.Unix(0, re.Timestamp),
			}
			if re.Attrs != "" {
				var attrMap map[string]string
				if json.Unmarshal([]byte(re.Attrs), &attrMap) == nil {
					for k, v := range attrMap {
						se.Attrs = append(se.Attrs, attribute.String(k, v))
					}
				}
			}
			spanEvents = append(spanEvents, se)
		}
	}

	if isError {
		if errMsg != "" {
			attrs = append(attrs, attribute.String("claude_code.error.message", errMsg))
		}
		attrs = append(attrs, attribute.Bool("claude_code.error.is_interrupt", isInterrupt))
		builder.CreateErrorSpan(toolCtx, spanName, startTime, endTime, attrs, errMsg, spanEvents...)
		_ = store.IncrementCounter(env.SessionID, "error_count")
	} else {
		builder.CreateSpan(toolCtx, spanName, startTime, endTime, attrs, spanEvents...)
	}

	_ = store.IncrementCounter(env.SessionID, "tool_count")

	durationMs := endTime.Sub(startTime).Milliseconds()

	// Record event on parent span timeline (prompt or subagent)
	if err := store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    tool.ParentSpanID,
		Name:      "tool.end",
		Timestamp: endTime.UnixNano(),
		Attrs:     fmt.Sprintf(`{"tool.name":"%s","duration_ms":"%d","success":"%v"}`, event.ToolName, durationMs, !isError),
	}); err != nil {
		debug.Log("failed to record tool event: %v", err)
	} else {
		debug.Log("recorded tool.end event on span %s", tool.ParentSpanID)
	}

	// Emit event
	eventName := "claude_code.tool.end"
	if isError {
		eventName = "claude_code.tool.error"
	}
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.tool.name"] = event.ToolName
	logAttrs["claude_code.tool.use_id"] = event.ToolUseID
	logAttrs["claude_code.tool.duration_ms"] = fmt.Sprintf("%d", durationMs)
	logAttrs["claude_code.tool.success"] = fmt.Sprintf("%v", !isError)
	if tool.FilePath != "" {
		fi := fileinfo.FromPath(tool.FilePath)
		logAttrs["claude_code.file.path"] = fi.Path
		logAttrs["claude_code.file.extension"] = fi.Extension
		if fi.Language != "" {
			logAttrs["claude_code.file.language"] = fi.Language
		}
	}
	if isError {
		if errMsg != "" {
			logAttrs["claude_code.error.message"] = errMsg
		}
		logAttrs["claude_code.error.is_interrupt"] = fmt.Sprintf("%v", isInterrupt)
	}
	provider.EmitEvent(eventName, sess.TraceID, tool.SpanID, logAttrs)

	// Emit metrics — build shared file attrs once
	var fi fileinfo.Info
	var fileMetricAttrs []attribute.KeyValue
	if tool.FilePath != "" {
		fi = fileinfo.FromPath(tool.FilePath)
		fileMetricAttrs = append(fileMetricAttrs,
			attribute.String("claude_code.file.extension", fi.Extension),
		)
		if fi.Language != "" {
			fileMetricAttrs = append(fileMetricAttrs,
				attribute.String("claude_code.file.language", fi.Language),
			)
		}
	}

	vcsAttrs := vcsMetricAttrs(env.Cwd)

	// tool.count
	toolCountAttrs := []attribute.KeyValue{
		attribute.String("claude_code.tool.name", event.ToolName),
		attribute.Bool("claude_code.tool.success", !isError),
		attribute.String("claude_code.session.cwd", env.Cwd),
	}
	toolCountAttrs = append(toolCountAttrs, fileMetricAttrs...)
	toolCountAttrs = append(toolCountAttrs, vcsAttrs...)
	provider.CounterAdd(ctx, "claude_code.tool.count", 1, toolCountAttrs...)

	// tool.duration
	toolDurationAttrs := []attribute.KeyValue{
		attribute.String("claude_code.tool.name", event.ToolName),
		attribute.Bool("claude_code.tool.success", !isError),
		attribute.String("claude_code.session.cwd", env.Cwd),
	}
	toolDurationAttrs = append(toolDurationAttrs, fileMetricAttrs...)
	toolDurationAttrs = append(toolDurationAttrs, vcsAttrs...)
	provider.HistogramRecord(ctx, "claude_code.tool.duration", float64(durationMs), toolDurationAttrs...)

	// error.count
	if isError {
		errorAttrs := []attribute.KeyValue{
			attribute.String("claude_code.tool.name", event.ToolName),
			attribute.Bool("claude_code.error.is_interrupt", isInterrupt),
		}
		errorAttrs = append(errorAttrs, vcsAttrs...)
		provider.CounterAdd(ctx, "claude_code.error.count", 1, errorAttrs...)
	}

	// lines_changed.count
	if linesAdded > 0 || linesRemoved > 0 {
		lineAttrs := []attribute.KeyValue{
			attribute.String("claude_code.session.cwd", env.Cwd),
		}
		lineAttrs = append(lineAttrs, fileMetricAttrs...)
		lineAttrs = append(lineAttrs, vcsAttrs...)

		if linesAdded > 0 {
			provider.CounterAdd(ctx, "claude_code.lines_changed.count", int64(linesAdded),
				append(lineAttrs, attribute.String("type", "added"))...,
			)
		}
		if linesRemoved > 0 {
			provider.CounterAdd(ctx, "claude_code.lines_changed.count", int64(linesRemoved),
				append(lineAttrs, attribute.String("type", "removed"))...,
			)
		}
	}

	debug.Log("tool end: session=%s tool=%s id=%s error=%v duration=%dms",
		env.SessionID, event.ToolName, event.ToolUseID, isError, durationMs)

	return store.DeleteTool(event.ToolUseID)
}

func computeLineDiff(event toolEvent, tool state.Tool, store *state.Store) (added, removed int) {
	switch event.ToolName {
	case "Edit":
		var input editInput
		if err := json.Unmarshal(event.ToolInput, &input); err == nil {
			return fileinfo.DiffLines(input.OldString, input.NewString)
		}
	case "Write":
		if tool.FileSnapshot != nil {
			// Compare snapshot (before) with new content from tool_input
			var input writeInput
			if err := json.Unmarshal(event.ToolInput, &input); err == nil {
				return fileinfo.DiffLines(string(tool.FileSnapshot), input.Content)
			}
		} else {
			// New file — all lines are added
			var input writeInput
			if err := json.Unmarshal(event.ToolInput, &input); err == nil {
				return fileinfo.CountLines(input.Content), 0
			}
		}
	}
	return 0, 0
}

func detectCommit(env payload.Envelope, tool state.Tool, store *state.Store, attrs *[]attribute.KeyValue) {
	preSHA, _ := store.GetCache("git_head_sha_pre_" + tool.ToolUseID)
	if preSHA == "" {
		return
	}

	postSHA := gitpkg.GetContext(env.Cwd).HeadSHA
	if postSHA != "" && postSHA != preSHA {
		*attrs = append(*attrs,
			attribute.String("claude_code.git.commit_sha", postSHA),
			attribute.Bool("claude_code.git.commit_detected", true),
		)
		_ = store.IncrementCounter(env.SessionID, "commit_count")
		debug.Log("commit detected: %s -> %s", preSHA, postSHA)
	}
}
