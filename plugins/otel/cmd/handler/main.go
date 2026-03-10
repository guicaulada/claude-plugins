package main

import (
	"io"
	"os"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/debug"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/dispatch"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/events"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
)

var registry = newRegistry()

func newRegistry() *dispatch.Registry {
	r := dispatch.New()
	r.Register("SessionStart", events.HandleSessionStart)
	r.Register("UserPromptSubmit", events.HandleUserPromptSubmit)
	r.Register("PreToolUse", events.HandlePreToolUse)
	r.Register("PostToolUse", events.HandlePostToolUse)
	r.Register("PostToolUseFailure", events.HandlePostToolUseFailure)
	r.Register("SubagentStart", events.HandleSubagentStart)
	r.Register("SubagentStop", events.HandleSubagentStop)
	r.Register("Stop", events.HandleStop)
	r.Register("PermissionRequest", events.HandlePermissionRequest)
	r.Register("Notification", events.HandleNotification)
	r.Register("TaskCompleted", events.HandleTaskCompleted)
	r.Register("InstructionsLoaded", events.HandleInstructionsLoaded)
	r.Register("ConfigChange", events.HandleConfigChange)
	r.Register("WorktreeCreate", events.HandleWorktreeCreate)
	r.Register("WorktreeRemove", events.HandleWorktreeRemove)
	r.Register("TeammateIdle", events.HandleTeammateIdle)
	r.Register("PreCompact", events.HandlePreCompact)
	r.Register("SessionEnd", events.HandleSessionEnd)
	return r
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			debug.Log("panic recovered: %v", r)
		}
		os.Exit(0)
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

	if err := registry.Dispatch(env); err != nil {
		debug.Log("handler error for %s: %v", env.HookEventName, err)
	}
}
