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
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

type userPromptEvent struct {
	Prompt string `json:"prompt"`
}

func HandleUserPromptSubmit(env payload.Envelope) error {
	var event userPromptEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse UserPromptSubmit event: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer store.Close()

	count, err := store.GetPromptCount(env.SessionID)
	if err != nil {
		return err
	}

	prompt := state.Prompt{
		SessionID:   env.SessionID,
		SpanID:      idgen.SpanID(),
		StartTime:   time.Now().UnixNano(),
		PromptIndex: count + 1,
	}

	if err := store.CreatePrompt(prompt); err != nil {
		return err
	}

	if err := store.IncrementCounter(env.SessionID, "prompt_count"); err != nil {
		return err
	}

	// Record event on session span timeline
	sess, _ := store.GetSession(env.SessionID)
	if sess.SessionID != "" {
		_ = store.RecordEvent(state.SpanEvent{
			SessionID: env.SessionID,
			SpanID:    sess.SpanID,
			Name:      "prompt.submit",
			Timestamp: prompt.StartTime,
			Attrs:     marshalAttrs(map[string]string{"prompt.index": fmt.Sprintf("%d", prompt.PromptIndex)}),
		})
	}

	// Emit event and metric
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProviderFromState(ctx, cfg, store)
	if err == nil {
		defer provider.Shutdown(ctx)

		sess, _ := store.GetSession(env.SessionID)
		logAttrs := commonLogAttrs(env)
		logAttrs["claude_code.prompt.index"] = fmt.Sprintf("%d", prompt.PromptIndex)
		if cfg.LogUserPrompts && event.Prompt != "" {
			logAttrs["claude_code.prompt.content"] = event.Prompt
		}
		provider.EmitEvent("claude_code.prompt.submit", sess.TraceID, prompt.SpanID, logAttrs)

		metricAttrs := []attribute.KeyValue{
			attribute.String("claude_code.session.cwd", env.Cwd),
		}
		metricAttrs = append(metricAttrs, vcsMetricAttrs(env.Cwd)...)
		provider.CounterAdd(ctx, "claude_code.prompt.count", 1, metricAttrs...)
	}

	debug.Log("prompt submit: session=%s prompt_index=%d", env.SessionID, prompt.PromptIndex)
	return nil
}
