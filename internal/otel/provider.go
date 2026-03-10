package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
)

const (
	ServiceName = "claude-code-otel-plugin"
	TracerName  = "claude-code-otel-plugin"
)

// Provider wraps the OTel TracerProvider and exposes a Tracer.
type Provider struct {
	tp     *sdktrace.TracerProvider
	tracer trace.Tracer
}

// NewProvider creates and configures the OTel TracerProvider.
func NewProvider(ctx context.Context, cfg config.Config) (*Provider, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(2 * time.Second),
	}

	if endpoint := cfg.OTelEndpoint(); endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	if cfg.OTelInsecure() {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	headers := cfg.OTelHeaders()
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			attribute.String("service.version", cfg.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(res),
	)

	return &Provider{
		tp:     tp,
		tracer: tp.Tracer(TracerName),
	}, nil
}

// Tracer returns the configured tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Shutdown flushes and shuts down the provider.
func (p *Provider) Shutdown(ctx context.Context) {
	if err := p.tp.Shutdown(ctx); err != nil {
		debug.Log("otel shutdown error: %v", err)
	}
}
