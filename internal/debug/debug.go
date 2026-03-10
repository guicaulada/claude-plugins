package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	enabled bool
	once    sync.Once
)

func isEnabled() bool {
	once.Do(func() {
		enabled = os.Getenv("OTEL_PLUGIN_DEBUG") == "1"
	})
	return enabled
}

func Log(format string, args ...any) {
	if !isEnabled() {
		return
	}

	dir := filepath.Join(os.TempDir(), "claude-code-otel-plugin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	f, err := os.OpenFile(filepath.Join(dir, "debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	ts := time.Now().UTC().Format(time.RFC3339Nano)
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(f, "%s %s\n", ts, msg)
}
