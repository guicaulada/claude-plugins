package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/log"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
	pluginotel "github.com/guicaulada/claude-plugins/plugins/otel/internal/otel"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/idgen"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/state"
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
	defer func() { _ = store.Close() }()

	// Increment prompt_count first so the index reflects the current prompt.
	// Using the persistent counter (not row count) avoids reset to 1 after
	// each prompt is deleted by HandleStop.
	if err := store.IncrementCounter(env.SessionID, "prompt_count"); err != nil {
		return err
	}

	count, err := store.GetCounter(env.SessionID, "prompt_count")
	if err != nil {
		return err
	}

	prompt := state.Prompt{
		SessionID:   env.SessionID,
		SpanID:      idgen.SpanID(),
		StartTime:   time.Now().UnixNano(),
		PromptIndex: int(count),
	}

	if err := store.CreatePrompt(prompt); err != nil {
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
	provider, err := newProvider(ctx, cfg)
	if err == nil {
		defer provider.Shutdown(ctx)

		sess, _ := store.GetSession(env.SessionID)
		logAttrs := commonLogAttrs(env)
		logAttrs = append(logAttrs, log.Int("claude_code.prompt.index", prompt.PromptIndex))
		if cfg.LogUserPrompts && event.Prompt != "" {
			logAttrs = append(logAttrs, log.String("claude_code.prompt.content", event.Prompt))
		}
		provider.EmitEvent("claude_code.prompt.submit", sess.TraceID, prompt.SpanID, logAttrs)

		metricAttrs := cwdMetricAttr(env.Cwd, cfg.IncludeHighCardinality)
		metricAttrs = append(metricAttrs, vcsMetricAttrs(env.Cwd, cfg.IncludeHighCardinality)...)
		provider.CounterAdd(ctx, pluginotel.MetricPrompts, 1, metricAttrs...)
	}

	debug.Log("prompt submit: session=%s prompt_index=%d", env.SessionID, prompt.PromptIndex)
	return nil
}
