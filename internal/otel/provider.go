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

// ProviderOption configures the OTel provider.
type ProviderOption func(*providerOptions)

type providerOptions struct {
	headers map[string]string
}

// WithHeaders passes pre-loaded headers (e.g., from state cache).
func WithHeaders(headers map[string]string) ProviderOption {
	return func(o *providerOptions) {
		o.headers = headers
	}
}

// NewProvider creates and configures the OTel TracerProvider.
// The exporter reads standard OTEL_EXPORTER_OTLP_* env vars automatically.
// Only plugin-specific overrides (OTEL_PLUGIN_EXPORTER_*) are set programmatically.
func NewProvider(ctx context.Context, cfg config.Config, opts ...ProviderOption) (*Provider, error) {
	var po providerOptions
	for _, o := range opts {
		o(&po)
	}

	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(2 * time.Second),
	}

	// Only override endpoint if plugin-specific env vars are set
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		exporterOpts = append(exporterOpts, otlptracehttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
		}
	}

	// Headers: pre-loaded > plugin override > otelHeadersHelper > SDK env var fallback
	if len(po.headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracehttp.WithHeaders(po.headers))
	} else if headers := cfg.PluginHeaders(); len(headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracehttp.WithHeaders(headers))
	} else if headers := config.LoadOTelHeaders(); len(headers) > 0 {
		exporterOpts = append(exporterOpts, otlptracehttp.WithHeaders(headers))
	}

	exp, err := otlptracehttp.New(ctx, exporterOpts...)
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
		sdktrace.WithIDGenerator(newFixedIDGenerator()),
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
