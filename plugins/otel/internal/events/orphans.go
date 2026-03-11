package events

import (
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	pluginotel "github.com/guicaulada/claude-plugins/plugins/otel/internal/otel"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/state"
)

// exportOrphanedTools exports and deletes any tools whose parent is parentSpanID,
// ending them at the given endTime. Returns the number of orphans exported.
func exportOrphanedTools(
	store *state.Store,
	builder *pluginotel.SpanBuilder,
	sessionID, traceID, parentSpanID string,
	endTime time.Time,
) int {
	tools, err := store.GetToolsByParent(sessionID, parentSpanID)
	if err != nil || len(tools) == 0 {
		return 0
	}

	debug.Log("exporting %d orphaned tool spans for parent %s", len(tools), parentSpanID)
	for _, tool := range tools {
		toolCtx, err := pluginotel.ChildContext(traceID, tool.ParentSpanID, tool.SpanID)
		if err != nil {
			continue
		}
		spanEvents := loadSpanEvents(store, sessionID, tool.SpanID)
		attrs := []attribute.KeyValue{
			attribute.String("claude_code.session.id", sessionID),
			attribute.String("claude_code.tool.name", tool.ToolName),
			attribute.String("claude_code.tool.use_id", tool.ToolUseID),
			attribute.Bool("claude_code.interrupted", true),
		}
		if tool.FilePath != "" {
			attrs = append(attrs, attribute.String("claude_code.file.path", tool.FilePath))
		}
		builder.CreateErrorSpan(toolCtx, "tool:"+tool.ToolName, time.Unix(0, tool.StartTime), endTime, attrs, "interrupted", spanEvents...)
		_ = store.DeleteTool(tool.ToolUseID)
		_ = store.IncrementCounter(sessionID, "interrupted_count")
	}
	return len(tools)
}

// exportOrphanedSubagents exports and deletes any subagents whose parent is parentSpanID,
// ending them at the given endTime. Returns the number of orphans exported.
func exportOrphanedSubagents(
	store *state.Store,
	builder *pluginotel.SpanBuilder,
	sessionID, traceID, parentSpanID string,
	endTime time.Time,
) int {
	agents, err := store.GetSubagentsByParent(sessionID, parentSpanID)
	if err != nil || len(agents) == 0 {
		return 0
	}

	debug.Log("exporting %d orphaned subagent spans for parent %s", len(agents), parentSpanID)
	for _, sa := range agents {
		saCtx, err := pluginotel.ChildContext(traceID, sa.ParentSpanID, sa.SpanID)
		if err != nil {
			continue
		}
		// Export any orphaned tools inside this subagent first
		exportOrphanedTools(store, builder, sessionID, traceID, sa.SpanID, endTime)

		spanEvents := loadSpanEvents(store, sessionID, sa.SpanID)
		attrs := []attribute.KeyValue{
			attribute.String("claude_code.session.id", sessionID),
			attribute.String("claude_code.agent.type", sa.AgentType),
			attribute.String("claude_code.agent.name", sa.AgentName),
			attribute.String("claude_code.agent.id", sa.AgentID),
			attribute.Bool("claude_code.interrupted", true),
		}
		builder.CreateErrorSpan(saCtx, "agent:"+sa.AgentType, time.Unix(0, sa.StartTime), endTime, attrs, "interrupted", spanEvents...)
		_ = store.DeleteSubagent(sa.AgentID)
		_ = store.IncrementCounter(sessionID, "interrupted_count")
	}
	return len(agents)
}

// loadSpanEvents reads recorded span events from the store.
func loadSpanEvents(store *state.Store, sessionID, spanID string) []pluginotel.SpanEvent {
	recorded, err := store.GetEvents(sessionID, spanID)
	if err != nil || len(recorded) == 0 {
		return nil
	}

	events := make([]pluginotel.SpanEvent, 0, len(recorded))
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
		events = append(events, se)
	}
	return events
}
