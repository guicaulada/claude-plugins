package events

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
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
		// Fallback to session span
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
	var toolInput struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(event.ToolInput, &toolInput); err == nil && toolInput.FilePath != "" {
		tool.FilePath = toolInput.FilePath
	}

	if err := store.CreateTool(tool); err != nil {
		return err
	}

	debug.Log("pre tool use: session=%s tool=%s id=%s parent=%s",
		env.SessionID, event.ToolName, event.ToolUseID, parentSpanID)
	return nil
}

func HandlePostToolUse(env payload.Envelope) error {
	return handleToolEnd(env, false, "")
}

func HandlePostToolUseFailure(env payload.Envelope) error {
	var event struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse PostToolUseFailure event: %v", err)
	}
	return handleToolEnd(env, true, event.Error)
}

func handleToolEnd(env payload.Envelope, isError bool, errMsg string) error {
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

	// Use ChildContext to preserve the stored span ID so subagent spans
	// that reference this tool as parent link correctly
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

	if tool.FilePath != "" {
		attrs = append(attrs, attribute.String("claude_code.file.path", tool.FilePath))
	}

	if isError {
		builder.CreateErrorSpan(toolCtx, spanName, startTime, endTime, attrs, errMsg)
		_ = store.IncrementCounter(env.SessionID, "error_count")
	} else {
		builder.CreateSpan(toolCtx, spanName, startTime, endTime, attrs)
	}

	_ = store.IncrementCounter(env.SessionID, "tool_count")

	debug.Log("tool end: session=%s tool=%s id=%s error=%v duration=%dms",
		env.SessionID, event.ToolName, event.ToolUseID, isError,
		endTime.Sub(startTime).Milliseconds())

	return store.DeleteTool(event.ToolUseID)
}
