package events

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Parent: prefer the most recent active Agent tool (direct parent),
	// fall back to current prompt, then session
	parentSpanID := ""
	agentTool, err := store.GetLatestToolByName(env.SessionID, "Agent")
	if err == nil && agentTool.ToolUseID != "" {
		parentSpanID = agentTool.SpanID
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

	// Use agent_type as fallback for empty agent_name
	agentName := event.AgentName
	if agentName == "" {
		agentName = event.AgentType
	}

	sa := state.Subagent{
		AgentID:      event.AgentID,
		SessionID:    env.SessionID,
		SpanID:       idgen.SpanID(),
		ParentSpanID: parentSpanID,
		StartTime:    time.Now().UnixNano(),
		AgentType:    event.AgentType,
		AgentName:    agentName,
	}

	if err := store.CreateSubagent(sa); err != nil {
		return err
	}

	_ = store.IncrementCounter(env.SessionID, "subagent_count")

	// Emit event and metric
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err == nil {
		defer provider.Shutdown(ctx)

		sess, _ := store.GetSession(env.SessionID)
		startLogAttrs := commonLogAttrs(env)
		startLogAttrs["claude_code.agent.type"] = event.AgentType
		startLogAttrs["claude_code.agent.name"] = agentName
		startLogAttrs["claude_code.agent.id"] = event.AgentID
		provider.EmitEvent("claude_code.agent.start", sess.TraceID, sa.SpanID, startLogAttrs)

		// Record event on parent span timeline (prompt)
		if err := store.RecordEvent(state.SpanEvent{
			SessionID: env.SessionID,
			SpanID:    sa.ParentSpanID,
			Name:      "agent.start",
			Timestamp: sa.StartTime,
			Attrs:     marshalAttrs(map[string]string{"agent.type": event.AgentType, "agent.name": agentName}),
		}); err != nil {
			debug.Log("failed to record agent.start event: %v", err)
		} else {
			debug.Log("recorded agent.start event on span %s", sa.ParentSpanID)
		}

		saMetricAttrs := []attribute.KeyValue{
			attribute.String("claude_code.agent.type", event.AgentType),
			attribute.String("claude_code.agent.name", agentName),
		}
		saMetricAttrs = append(saMetricAttrs, vcsMetricAttrs(env.Cwd)...)
		provider.CounterAdd(ctx, "claude_code.subagent.count", 1, saMetricAttrs...)
	}

	debug.Log("subagent start: session=%s agent=%s type=%s id=%s",
		env.SessionID, agentName, event.AgentType, event.AgentID)
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
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer provider.Shutdown(ctx)

	// Use ChildContext so the subagent span gets the exact span ID from state
	// (which tool spans inside this subagent reference as their parent)
	saCtx, err := pluginotel.ChildContext(sess.TraceID, sa.ParentSpanID, sa.SpanID)
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

	// VCS enrichment
	attrs = append(attrs, vcsAttributes(env.Cwd, env.SessionID, store)...)

	// Load recorded events for this subagent span (tool calls within it)
	var spanEvents []pluginotel.SpanEvent
	if recorded, err := store.GetEvents(env.SessionID, sa.SpanID); err == nil {
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

	builder.CreateSpan(saCtx, spanName, startTime, endTime, attrs, spanEvents...)

	durationMs := endTime.Sub(startTime).Milliseconds()

	// Record agent.stop event on parent span (Agent tool span)
	if err := store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    sa.ParentSpanID,
		Name:      "agent.stop",
		Timestamp: endTime.UnixNano(),
		Attrs:     marshalAttrs(map[string]string{"agent.type": sa.AgentType, "agent.name": sa.AgentName, "duration_ms": fmt.Sprintf("%d", durationMs)}),
	}); err != nil {
		debug.Log("failed to record agent.stop event: %v", err)
	} else {
		debug.Log("recorded agent.stop event on parent span %s", sa.ParentSpanID)
	}

	// Emit event
	stopLogAttrs := commonLogAttrs(env)
	stopLogAttrs["claude_code.agent.type"] = sa.AgentType
	stopLogAttrs["claude_code.agent.name"] = sa.AgentName
	stopLogAttrs["claude_code.agent.id"] = sa.AgentID
	stopLogAttrs["claude_code.agent.duration_ms"] = fmt.Sprintf("%d", durationMs)
	provider.EmitEvent("claude_code.agent.stop", sess.TraceID, sa.SpanID, stopLogAttrs)

	// Emit metric
	saDurationAttrs := []attribute.KeyValue{
		attribute.String("claude_code.agent.type", sa.AgentType),
		attribute.String("claude_code.agent.name", sa.AgentName),
	}
	saDurationAttrs = append(saDurationAttrs, vcsMetricAttrs(env.Cwd)...)
	provider.HistogramRecord(ctx, "claude_code.subagent.duration", float64(durationMs), saDurationAttrs...)

	debug.Log("subagent stop: session=%s agent=%s duration=%dms",
		env.SessionID, sa.AgentType, durationMs)

	return store.DeleteSubagent(event.AgentID)
}
