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

// OTelEndpoint returns the OTLP endpoint, preferring plugin-specific override.
func (c Config) OTelEndpoint() string {
	if v := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return stripScheme(v)
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return stripScheme(v)
	}
	return ""
}

// OTelInsecure returns true if the endpoint uses HTTP (not HTTPS).
func (c Config) OTelInsecure() bool {
	endpoint := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}
	return strings.HasPrefix(endpoint, "http://")
}

// OTelHeaders returns the OTLP headers as a map, preferring plugin-specific override.
func (c Config) OTelHeaders() map[string]string {
	raw := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_HEADERS")
	if raw == "" {
		raw = os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	}
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

// stripScheme removes http:// or https:// prefix for the OTLP exporter.
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
