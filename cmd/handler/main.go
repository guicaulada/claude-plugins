package main

import (
	"io"
	"os"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/dispatch"
	"github.com/guicaulada/claude-code-otel-plugin/internal/events"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
)

func main() {
	code := 0
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic recovered: %v", r)
		}
		os.Exit(code)
	}()

	run()
}

func run() {
	cfg := config.Load()
	if !cfg.Enabled {
		return
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		debug.Log("failed to read stdin: %v", err)
		return
	}

	if len(input) == 0 {
		debug.Log("empty stdin, nothing to process")
		return
	}

	env, err := payload.Parse(input)
	if err != nil {
		debug.Log("failed to parse payload: %v", err)
		return
	}

	registry := dispatch.New()
	registerHandlers(registry)

	if err := registry.Dispatch(env); err != nil {
		debug.Log("handler error for %s: %v", env.HookEventName, err)
	}
}

func registerHandlers(r *dispatch.Registry) {
	r.Register("SessionStart", events.HandleSessionStart)
	r.Register("UserPromptSubmit", events.HandleUserPromptSubmit)
	r.Register("PreToolUse", events.HandlePreToolUse)
	r.Register("PostToolUse", events.HandlePostToolUse)
	r.Register("PostToolUseFailure", events.HandlePostToolUseFailure)
	r.Register("SubagentStart", events.HandleSubagentStart)
	r.Register("SubagentStop", events.HandleSubagentStop)
	r.Register("Stop", events.HandleStop)
	r.Register("SessionEnd", events.HandleSessionEnd)
}
