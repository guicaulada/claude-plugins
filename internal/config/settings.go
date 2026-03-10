package config

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
)

type settingsFile struct {
	OTelHeadersHelper string `json:"otelHeadersHelper"`
}

// LoadOTelHeaders resolves the otelHeadersHelper from Claude Code settings
// hierarchy (local > project > user), runs the script, and returns headers.
func LoadOTelHeaders() map[string]string {
	helper := findHeadersHelper()
	if helper == "" {
		debug.Log("no otelHeadersHelper found in settings")
		return nil
	}

	debug.Log("running otelHeadersHelper: %s", helper)
	headers, err := runHeadersHelper(helper)
	if err != nil {
		debug.Log("otelHeadersHelper failed: %v", err)
		return nil
	}

	debug.Log("otelHeadersHelper returned %d headers", len(headers))
	return headers
}

func findHeadersHelper() string {
	// Check settings in precedence order: local > project > user
	paths := settingsPaths()
	for _, p := range paths {
		if helper := readHelperFromSettings(p); helper != "" {
			return helper
		}
	}
	return ""
}

func settingsPaths() []string {
	var paths []string

	// Project-local settings
	cwd, _ := os.Getwd()
	if cwd != "" {
		paths = append(paths, filepath.Join(cwd, ".claude", "settings.local.json"))
		paths = append(paths, filepath.Join(cwd, ".claude", "settings.json"))
	}

	// User settings
	home, _ := os.UserHomeDir()
	if home != "" {
		paths = append(paths, filepath.Join(home, ".claude", "settings.json"))
	}

	return paths
}

func readHelperFromSettings(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var s settingsFile
	if err := json.Unmarshal(data, &s); err != nil {
		return ""
	}

	return s.OTelHeadersHelper
}

func runHeadersHelper(script string) (map[string]string, error) {
	cmd := exec.Command("sh", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil, nil
	}

	// The script outputs JSON: {"Header-Name": "value", ...}
	var headers map[string]string
	if err := json.Unmarshal([]byte(raw), &headers); err != nil {
		// Fallback: try "Name: Value" format (one per line)
		headers = make(map[string]string)
		for _, line := range strings.Split(raw, "\n") {
			k, v, ok := strings.Cut(line, ":")
			if ok {
				headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}

	return headers, nil
}
