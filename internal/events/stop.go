package events

import (
	"context"
	"encoding/json"
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

	// Load recorded events for this prompt span (tool calls, agent starts)
	var spanEvents []pluginotel.SpanEvent
	if recorded, err := store.GetEvents(env.SessionID, prompt.SpanID); err == nil {
		debug.Log("stop: loaded %d events for prompt span %s", len(recorded), prompt.SpanID)
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

	builder.CreateSpan(promptCtx, "prompt", startTime, endTime, attrs, spanEvents...)

	durationMs := endTime.Sub(startTime).Milliseconds()

	// Record event on session span timeline
	_ = store.RecordEvent(state.SpanEvent{
		SessionID: env.SessionID,
		SpanID:    sess.SpanID,
		Name:      "prompt.stop",
		Timestamp: endTime.UnixNano(),
		Attrs:     fmt.Sprintf(`{"prompt.index":"%d","duration_ms":"%d"}`, prompt.PromptIndex, durationMs),
	})

	// Emit event
	logAttrs := commonLogAttrs(env)
	logAttrs["claude_code.prompt.index"] = fmt.Sprintf("%d", prompt.PromptIndex)
	logAttrs["claude_code.prompt.duration_ms"] = fmt.Sprintf("%d", durationMs)
	provider.EmitEvent("claude_code.prompt.stop", sess.TraceID, prompt.SpanID, logAttrs)

	// Emit metric
	metricAttrs := []attribute.KeyValue{
		attribute.String("claude_code.session.cwd", env.Cwd),
	}
	metricAttrs = append(metricAttrs, vcsMetricAttrs(env.Cwd)...)
	provider.HistogramRecord(ctx, "claude_code.prompt.duration", float64(durationMs), metricAttrs...)

	debug.Log("stop: exported prompt span session=%s index=%d duration=%dms",
		env.SessionID, prompt.PromptIndex, durationMs)

	// Clean up the prompt from state
	return store.DeletePrompt(prompt.ID)
}
