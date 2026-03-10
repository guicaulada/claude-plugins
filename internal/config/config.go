package config

import (
	"net/url"
	"os"
	"strings"
)

// Version is set at build time via ldflags.
var Version = "dev"

type Config struct {
	Enabled bool
	Debug   bool
	Version string
}

func Load() Config {
	pluginEnabled := os.Getenv("OTEL_PLUGIN_ENABLED")
	telemetryEnabled := os.Getenv("CLAUDE_CODE_ENABLE_TELEMETRY")
	debug := os.Getenv("OTEL_PLUGIN_DEBUG") == "1"

	enabled := resolveEnabled(pluginEnabled, telemetryEnabled)

	return Config{
		Enabled: enabled,
		Debug:   debug,
		Version: Version,
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
	switch pluginEnabled {
	case "1", "true":
		return true
	case "0", "false":
		return false
	}
	return telemetryEnabled == "1"
}

// stripScheme removes http:// or https:// prefix for the OTLP exporter WithEndpoint.
func stripScheme(endpoint string) string {
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		return u.Host + u.Path
	}
	return endpoint
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
