package main

import (
	"io"
	"os"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	"github.com/guicaulada/claude-code-otel-plugin/internal/dispatch"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic recovered: %v", r)
		}
		os.Exit(0)
	}()

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
	r.Register("SessionStart", handleSessionStart)
	r.Register("SessionEnd", handleSessionEnd)
}

func handleSessionStart(env payload.Envelope) error {
	debug.Log("session start: %s (cwd: %s, mode: %s)", env.SessionID, env.Cwd, env.PermissionMode)
	return nil
}

func handleSessionEnd(env payload.Envelope) error {
	debug.Log("session end: %s", env.SessionID)
	return nil
}
