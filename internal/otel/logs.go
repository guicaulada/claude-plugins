package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/log"
)

// EmitEvent emits an OTel log record with trace correlation.
// The traceIDHex and spanIDHex are injected into the context so
// the SDK attaches them to the log record automatically.
func (p *Provider) EmitEvent(name string, traceIDHex, spanIDHex string, attrs map[string]string) {
	ctx := context.Background()

	// Inject trace context for Loki→Tempo correlation
	if traceIDHex != "" && spanIDHex != "" {
		if logCtx, err := ParentContext(traceIDHex, spanIDHex); err == nil {
			ctx = logCtx
		}
	}

	var record log.Record
	record.SetTimestamp(time.Now())
	record.SetEventName(name)
	record.SetBody(log.StringValue(name))
	record.SetSeverity(log.SeverityInfo)
	record.SetSeverityText("INFO")

	kvs := make([]log.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		kvs = append(kvs, log.String(k, v))
	}
	record.AddAttributes(kvs...)

	p.logger.Emit(ctx, record)
}
