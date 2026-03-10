package config

import (
	"testing"
)

func TestResolveEnabled(t *testing.T) {
	tests := []struct {
		name             string
		pluginEnabled    string
		telemetryEnabled string
		want             bool
	}{
		{"plugin explicitly enabled", "1", "0", true},
		{"plugin explicitly enabled true", "true", "0", true},
		{"plugin explicitly disabled", "0", "1", false},
		{"plugin explicitly disabled false", "false", "1", false},
		{"fallback to telemetry enabled", "", "1", true},
		{"fallback to telemetry disabled", "", "0", false},
		{"both unset", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveEnabled(tt.pluginEnabled, tt.telemetryEnabled)
			if got != tt.want {
				t.Errorf("resolveEnabled(%q, %q) = %v, want %v",
					tt.pluginEnabled, tt.telemetryEnabled, got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_ENABLED", "1")
	t.Setenv("OTEL_PLUGIN_DEBUG", "1")
	t.Setenv("OTEL_LOG_USER_PROMPTS", "1")
	t.Setenv("OTEL_LOG_TOOL_DETAILS", "1")
	t.Setenv("OTEL_METRICS_INCLUDE_VERSION", "true")

	cfg := Load()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if !cfg.Debug {
		t.Error("expected Debug to be true")
	}
	if !cfg.LogUserPrompts {
		t.Error("expected LogUserPrompts to be true")
	}
	if !cfg.LogToolDetails {
		t.Error("expected LogToolDetails to be true")
	}
	if !cfg.IncludeVersion {
		t.Error("expected IncludeVersion to be true")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_ENABLED", "")
	t.Setenv("CLAUDE_CODE_ENABLE_TELEMETRY", "")
	t.Setenv("OTEL_PLUGIN_DEBUG", "")
	t.Setenv("OTEL_LOG_USER_PROMPTS", "")
	t.Setenv("OTEL_LOG_TOOL_DETAILS", "")
	t.Setenv("OTEL_METRICS_INCLUDE_VERSION", "")

	cfg := Load()
	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.LogUserPrompts {
		t.Error("expected LogUserPrompts to be false by default")
	}
	if cfg.LogToolDetails {
		t.Error("expected LogToolDetails to be false by default")
	}
	if cfg.IncludeVersion {
		t.Error("expected IncludeVersion to be false by default")
	}
}

func TestPluginEndpoint(t *testing.T) {
	cfg := Config{}

	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")


	endpoint := cfg.PluginEndpoint()
	if endpoint != "localhost:4318" {
		t.Errorf("PluginEndpoint = %q, want %q", endpoint, "localhost:4318")
	}
}

func TestPluginEndpointEmpty(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT", "")

	cfg := Config{}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		t.Errorf("PluginEndpoint should be empty, got %q", endpoint)
	}
}

func TestPluginInsecure(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")


	cfg := Config{}
	if !cfg.PluginInsecure() {
		t.Error("expected insecure for http:// endpoint")
	}
}

func TestPluginInsecureHTTPS(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_ENDPOINT", "https://example.com")


	cfg := Config{}
	if cfg.PluginInsecure() {
		t.Error("expected secure for https:// endpoint")
	}
}

func TestPluginHeaders(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_HEADERS", "Authorization=Bearer token,X-Key=value")


	cfg := Config{}
	headers := cfg.PluginHeaders()
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("Authorization = %q", headers["Authorization"])
	}
	if headers["X-Key"] != "value" {
		t.Errorf("X-Key = %q", headers["X-Key"])
	}
}

func TestPluginProtocol(t *testing.T) {
	t.Setenv("OTEL_PLUGIN_EXPORTER_OTLP_PROTOCOL", "http/json")


	cfg := Config{}
	if p := cfg.PluginProtocol(); p != "http/json" {
		t.Errorf("PluginProtocol = %q, want %q", p, "http/json")
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{"key=value", map[string]string{"key": "value"}},
		{"a=1,b=2", map[string]string{"a": "1", "b": "2"}},
		{" key = value ", map[string]string{"key": "value"}},
		{"", map[string]string{}},
	}

	for _, tt := range tests {
		got := parseHeaders(tt.input)
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("parseHeaders(%q)[%q] = %q, want %q", tt.input, k, got[k], v)
			}
		}
	}
}

func TestStripScheme(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://localhost:4318", "localhost:4318"},
		{"https://example.com/otlp", "example.com/otlp"},
		{"localhost:4318", "localhost:4318"},
	}

	for _, tt := range tests {
		got := stripScheme(tt.input)
		if got != tt.want {
			t.Errorf("stripScheme(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
