package events

import (
	"strings"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
)

// sanitizeToolName redacts MCP and Skill tool names when OTEL_LOG_TOOL_DETAILS is not enabled.
// MCP tools (mcp__<server>__<tool>) become "mcp_tool".
// Skill tool becomes "Skill".
// All other tool names pass through unchanged.
func sanitizeToolName(toolName string, cfg config.Config) string {
	if cfg.LogToolDetails {
		return toolName
	}

	if strings.HasPrefix(toolName, "mcp__") {
		return "mcp_tool"
	}

	return toolName
}
