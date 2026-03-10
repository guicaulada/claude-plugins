package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
)

// headerCache is stored as a JSON file shared across all sessions.
type headerCache struct {
	Headers   map[string]string `json:"headers"`
	Timestamp int64             `json:"timestamp"` // UnixMilli
}

// headerCachePath returns the shared cache file path.
func headerCachePath() string {
	return filepath.Join(os.TempDir(), "claude-code-otel-plugin", "headers.json")
}

// debounceInterval returns the header refresh interval from env var or default (29 min).
func debounceInterval() time.Duration {
	if ms := os.Getenv("CLAUDE_CODE_OTEL_HEADERS_HELPER_DEBOUNCE_MS"); ms != "" {
		if v, err := strconv.ParseInt(ms, 10, 64); err == nil {
			return time.Duration(v) * time.Millisecond
		}
	}
	return 29 * time.Minute // Claude Code default
}

// LoadCachedHeaders returns cached headers if fresh, or loads from otelHeadersHelper
// and updates the shared cache. The cache is shared across all sessions.
func LoadCachedHeaders() map[string]string {
	cachePath := headerCachePath()
	debounce := debounceInterval()

	// Try reading shared cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var cache headerCache
		if err := json.Unmarshal(data, &cache); err == nil && len(cache.Headers) > 0 {
			age := time.Since(time.UnixMilli(cache.Timestamp))
			if age < debounce {
				debug.Log("using shared cached headers (age: %s, debounce: %s)", age.Round(time.Second), debounce)
				return cache.Headers
			}
			debug.Log("shared header cache expired (age: %s, debounce: %s)", age.Round(time.Second), debounce)
		}
	}

	// Cache miss or stale — load from helper
	headers := LoadOTelHeaders()
	if len(headers) == 0 {
		return nil
	}

	// Write to shared cache
	cache := headerCache{
		Headers:   headers,
		Timestamp: time.Now().UnixMilli(),
	}
	if data, err := json.Marshal(cache); err == nil {
		dir := filepath.Dir(cachePath)
		_ = os.MkdirAll(dir, 0o755)
		if err := os.WriteFile(cachePath, data, 0o600); err != nil {
			debug.Log("failed to write shared header cache: %v", err)
		} else {
			debug.Log("wrote shared header cache (%d headers)", len(headers))
		}
	}

	return headers
}
