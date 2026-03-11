package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/log"
)

// EmitEvent emits an OTel log record with trace correlation.
// The traceIDHex and spanIDHex are injected into the context so
// the SDK attaches them to the log record automatically.
// An optional body provides a human-readable display message; if empty,
// the event name is used as the body.
func (p *Provider) EmitEvent(name string, traceIDHex, spanIDHex string, attrs []log.KeyValue, body ...string) {
	ctx := context.Background()

	// Inject trace context for Loki→Tempo correlation
	if traceIDHex != "" && spanIDHex != "" {
		if logCtx, err := ParentContext(traceIDHex, spanIDHex); err == nil {
			ctx = logCtx
		}
	}

	sev, sevText := eventSeverity(name)

	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetEventName(name)
	bodyText := name
	if len(body) > 0 && body[0] != "" {
		bodyText = body[0]
	}
	record.SetBody(log.StringValue(bodyText))
	record.SetSeverity(sev)
	record.SetSeverityText(sevText)
	record.AddAttributes(attrs...)

	p.logger.Emit(ctx, record)
}

// eventSeverity maps event names to appropriate OTel severity levels.
func eventSeverity(name string) (log.Severity, string) {
	switch name {
	// Debug: lifecycle start events, informational loading
	case "claude_code.tool.start",
		"claude_code.agent.start",
		"claude_code.instructions.loaded",
		"claude_code.config.change",
		"claude_code.worktree.create",
		"claude_code.worktree.remove",
		"claude_code.compact":
		return log.SeverityDebug, "DEBUG"

	// Warn: permission gates, teammate waiting
	case "claude_code.permission.request",
		"claude_code.teammate.idle":
		return log.SeverityWarn, "WARN"

	// Error: tool failures
	case "claude_code.tool.error":
		return log.SeverityError, "ERROR"

	// Info: everything else (session start/end, prompt, tool end, notification, task)
	default:
		return log.SeverityInfo, "INFO"
	}
}
