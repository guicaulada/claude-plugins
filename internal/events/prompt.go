package events

import (
	"encoding/json"
	"time"

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

	debug.Log("prompt submit: session=%s prompt_index=%d", env.SessionID, prompt.PromptIndex)
	return nil
}
