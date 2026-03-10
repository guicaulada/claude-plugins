package config

import (
	"os"
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
	os.Setenv("OTEL_PLUGIN_ENABLED", "1")
	os.Setenv("OTEL_PLUGIN_DEBUG", "1")
	defer os.Unsetenv("OTEL_PLUGIN_ENABLED")
	defer os.Unsetenv("OTEL_PLUGIN_DEBUG")

	cfg := Load()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if !cfg.Debug {
		t.Error("expected Debug to be true")
	}
}
