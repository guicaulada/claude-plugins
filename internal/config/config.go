package config

import "os"

type Config struct {
	Enabled bool
	Debug   bool
}

func Load() Config {
	pluginEnabled := os.Getenv("OTEL_PLUGIN_ENABLED")
	telemetryEnabled := os.Getenv("CLAUDE_CODE_ENABLE_TELEMETRY")
	debug := os.Getenv("OTEL_PLUGIN_DEBUG") == "1"

	enabled := resolveEnabled(pluginEnabled, telemetryEnabled)

	return Config{
		Enabled: enabled,
		Debug:   debug,
	}
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
