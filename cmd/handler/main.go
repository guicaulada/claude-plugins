package main

import (
	"io"
	"os"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
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

	debug.Log("received %d bytes from stdin", len(input))
}
