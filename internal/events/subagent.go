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

type subagentStartEvent struct {
	AgentName string `json:"agent_name"`
	AgentType string `json:"agent_type"`
	AgentID   string `json:"agent_id"`
}

type subagentStopEvent struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
}

func HandleSubagentStart(env payload.Envelope) error {
	var event subagentStartEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse SubagentStart event: %v", err)
		return nil
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	// Parent is the current prompt span
	parentSpanID := ""
	prompt, err := store.GetCurrentPrompt(env.SessionID)
	if err == nil && prompt.SessionID != "" {
		parentSpanID = prompt.SpanID
	}
	if parentSpanID == "" {
		sess, err := store.GetSession(env.SessionID)
		if err == nil && sess.SessionID != "" {
			parentSpanID = sess.SpanID
		}
	}

	sa := state.Subagent{
		AgentID:      event.AgentID,
		SessionID:    env.SessionID,
		SpanID:       idgen.SpanID(),
		ParentSpanID: parentSpanID,
		StartTime:    time.Now().UnixNano(),
		AgentType:    event.AgentType,
		AgentName:    event.AgentName,
	}

	if err := store.CreateSubagent(sa); err != nil {
		return err
	}

	_ = store.IncrementCounter(env.SessionID, "subagent_count")

	debug.Log("subagent start: session=%s agent=%s type=%s id=%s",
		env.SessionID, event.AgentName, event.AgentType, event.AgentID)
	return nil
}

func HandleSubagentStop(env payload.Envelope) error {
	var event subagentStopEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse SubagentStop event: %v", err)
		return nil
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	sa, err := store.GetSubagent(event.AgentID)
	if err != nil || sa.AgentID == "" {
		debug.Log("subagent stop: no state for agent_id=%s", event.AgentID)
		return err
	}

	sess, err := store.GetSession(env.SessionID)
	if err != nil || sess.SessionID == "" {
		debug.Log("subagent stop: no session state for %s", env.SessionID)
		return err
	}

	// Export the subagent span
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProviderFromState(ctx, cfg, store)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	parentCtx, err := pluginotel.ParentContext(sess.TraceID, sa.ParentSpanID)
	if err != nil {
		return err
	}

	builder := pluginotel.NewSpanBuilder(provider.Tracer())
	startTime := time.Unix(0, sa.StartTime)
	endTime := time.Now()

	spanName := "agent:" + sa.AgentType
	attrs := []attribute.KeyValue{
		attribute.String("claude_code.session.id", env.SessionID),
		attribute.String("claude_code.agent.type", sa.AgentType),
		attribute.String("claude_code.agent.name", sa.AgentName),
		attribute.String("claude_code.agent.id", sa.AgentID),
		attribute.String("claude_code.permission_mode", env.PermissionMode),
	}

	builder.CreateSpan(parentCtx, spanName, startTime, endTime, attrs)

	debug.Log("subagent stop: session=%s agent=%s duration=%dms",
		env.SessionID, sa.AgentType,
		endTime.Sub(startTime).Milliseconds())

	return store.DeleteSubagent(event.AgentID)
}
