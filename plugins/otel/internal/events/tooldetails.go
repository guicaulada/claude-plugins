package events

import (
	"encoding/json"

	"go.opentelemetry.io/otel/log"
)

// addToolInputDetails extracts safe details from tool_input and adds them
// to log attributes. Only called when OTEL_LOG_TOOL_DETAILS=1.
func addToolInputDetails(attrs *[]log.KeyValue, toolName string, toolInput json.RawMessage) {
	if len(toolInput) == 0 {
		return
	}

	switch toolName {
	case "Bash":
		var input struct {
			Command     string `json:"command"`
			Description string `json:"description"`
			Timeout     int    `json:"timeout"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Command != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.command", input.Command))
			}
			if input.Description != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.description", input.Description))
			}
		}

	case "Edit":
		var input struct {
			FilePath   string `json:"file_path"`
			ReplaceAll bool   `json:"replace_all"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.FilePath != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.file_path", input.FilePath))
			}
		}

	case "Write":
		var input struct {
			FilePath string `json:"file_path"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.FilePath != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.file_path", input.FilePath))
			}
		}

	case "Read":
		var input struct {
			FilePath string `json:"file_path"`
			Offset   int    `json:"offset"`
			Limit    int    `json:"limit"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.FilePath != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.file_path", input.FilePath))
			}
		}

	case "Glob":
		var input struct {
			Pattern string `json:"pattern"`
			Path    string `json:"path"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Pattern != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.pattern", input.Pattern))
			}
			if input.Path != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.path", input.Path))
			}
		}

	case "Grep":
		var input struct {
			Pattern string `json:"pattern"`
			Path    string `json:"path"`
			Glob    string `json:"glob"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Pattern != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.pattern", input.Pattern))
			}
			if input.Path != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.path", input.Path))
			}
			if input.Glob != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.glob", input.Glob))
			}
		}

	case "Agent":
		var input struct {
			Description  string `json:"description"`
			SubagentType string `json:"subagent_type"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Description != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.description", input.Description))
			}
			if input.SubagentType != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.subagent_type", input.SubagentType))
			}
		}

	case "WebFetch":
		var input struct {
			URL string `json:"url"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.URL != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.url", input.URL))
			}
		}

	case "WebSearch":
		var input struct {
			Query string `json:"query"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Query != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.query", input.Query))
			}
		}

	case "Skill":
		var input struct {
			Skill string `json:"skill"`
			Args  string `json:"args"`
		}
		if json.Unmarshal(toolInput, &input) == nil {
			if input.Skill != "" {
				*attrs = append(*attrs, log.String("claude_code.tool.input.skill", input.Skill))
			}
		}
	}
}
