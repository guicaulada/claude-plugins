package config

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	Enabled              bool
	Debug                bool
	Version              string
	LogUserPrompts       bool
	LogToolDetails       bool
	IncludeHighCardinality bool
}

func Load() Config {
	pluginEnabled := os.Getenv("OTEL_PLUGIN_ENABLED")
	telemetryEnabled := os.Getenv("CLAUDE_CODE_ENABLE_TELEMETRY")
	debug := parseBool(os.Getenv("OTEL_PLUGIN_DEBUG"))

	enabled := resolveEnabled(pluginEnabled, telemetryEnabled)

	return Config{
		Enabled:        enabled,
		Debug:          debug,
		Version:        pluginVersion(),
		LogUserPrompts: parseBool(os.Getenv("OTEL_LOG_USER_PROMPTS")),
		LogToolDetails: parseBool(os.Getenv("OTEL_LOG_TOOL_DETAILS")),
		IncludeHighCardinality: parseBool(os.Getenv("OTEL_PLUGIN_METRICS_INCLUDE_HIGH_CARDINALITY")),
	}
}

// PluginEndpoint returns the plugin-specific OTLP endpoint override (host:port only).
// Returns empty if not set — the SDK will use OTEL_EXPORTER_OTLP_ENDPOINT automatically.
func (c Config) PluginEndpoint() string {
	if v := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return stripScheme(v)
	}
	return ""
}

// PluginInsecure returns true if the plugin-specific endpoint uses HTTP.
func (c Config) PluginInsecure() bool {
	return strings.HasPrefix(os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT"), "http://")
}

// PluginProtocol returns the plugin-specific OTLP protocol override.
// Returns empty if not set — the SDK will use OTEL_EXPORTER_OTLP_PROTOCOL automatically.
func (c Config) PluginProtocol() string {
	return os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_PROTOCOL")
}

// PluginHeaders returns plugin-specific OTLP header overrides.
// Returns nil if not set — the SDK will use OTEL_EXPORTER_OTLP_HEADERS automatically.
func (c Config) PluginHeaders() map[string]string {
	raw := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_HEADERS")
	if raw == "" {
		return nil
	}
	return parseHeaders(raw)
}

func resolveEnabled(pluginEnabled, telemetryEnabled string) bool {
	switch strings.ToLower(pluginEnabled) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	}
	return parseBool(telemetryEnabled)
}

// stripScheme removes http:// or https:// prefix for the OTLP exporter WithEndpoint.
func stripScheme(endpoint string) string {
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		return u.Host + u.Path
	}
	return endpoint
}

// parseBool accepts common truthy values: "1", "true", "yes".
func parseBool(s string) bool {
	switch strings.ToLower(s) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

// pluginVersion reads the version from .claude-plugin/plugin.json relative
// to the binary's location. This ensures the version always matches the
// plugin.json that release-please manages, regardless of when the binary
// was built.
var (
	pluginVersionOnce  sync.Once
	pluginVersionValue string
)

func pluginVersion() string {
	pluginVersionOnce.Do(func() {
		pluginVersionValue = "dev"
		exe, err := os.Executable()
		if err != nil {
			return
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return
		}
		// Binary is at <plugin_root>/bin/handler-*, plugin.json is at <plugin_root>/.claude-plugin/plugin.json
		pluginRoot := filepath.Dir(filepath.Dir(exe))
		data, err := os.ReadFile(filepath.Join(pluginRoot, ".claude-plugin", "plugin.json"))
		if err != nil {
			return
		}
		var meta struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &meta) == nil && meta.Version != "" {
			pluginVersionValue = meta.Version
		}
	})
	return pluginVersionValue
}

// parseHeaders parses "key=value,key=value" format into a map.
func parseHeaders(raw string) map[string]string {
	headers := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if ok {
			headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return headers
}
