package otel

import (
	"context"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
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
	MeterName   = "claude-code-otel-plugin"
	LoggerName  = "claude-code-otel-plugin"
)

// Provider wraps OTel TracerProvider, MeterProvider, and LoggerProvider.
type Provider struct {
	tp     *sdktrace.TracerProvider
	mp     *sdkmetric.MeterProvider
	lp     *sdklog.LoggerProvider
	tracer trace.Tracer
	logger otellog.Logger
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

// NewProvider creates and configures OTel TracerProvider, MeterProvider,
// and LoggerProvider. The exporters read standard OTEL_EXPORTER_OTLP_*
// env vars automatically. Only plugin-specific overrides are set programmatically.
func NewProvider(ctx context.Context, cfg config.Config, opts ...ProviderOption) (*Provider, error) {
	var po providerOptions
	for _, o := range opts {
		o(&po)
	}

	// If we have pre-loaded headers (from otelHeadersHelper cache) and
	// OTEL_EXPORTER_OTLP_HEADERS is not set, set it as env var so all
	// exporters (trace, metric, log) pick it up consistently.
	if len(po.headers) > 0 && os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") == "" {
		var pairs []string
		for k, v := range po.headers {
			pairs = append(pairs, k+"="+v)
		}
		os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", strings.Join(pairs, ","))
		debug.Log("set OTEL_EXPORTER_OTLP_HEADERS from cached headers (%d pairs)", len(po.headers))
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tp, err := newTracerProvider(ctx, cfg, po, res)
	if err != nil {
		return nil, err
	}

	mp, err := newMeterProvider(ctx, cfg, po, res)
	if err != nil {
		tp.Shutdown(ctx)
		return nil, err
	}

	lp, err := newLoggerProvider(ctx, cfg, po, res)
	if err != nil {
		tp.Shutdown(ctx)
		mp.Shutdown(ctx)
		return nil, err
	}

	return &Provider{
		tp:     tp,
		mp:     mp,
		lp:     lp,
		tracer: tp.Tracer(TracerName),
		logger: lp.Logger(LoggerName),
	}, nil
}

// Tracer returns the configured tracer.
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Meter returns the configured meter.
func (p *Provider) Meter() otelmetric.Meter {
	return p.mp.Meter(MeterName)
}

// Logger returns the configured logger.
func (p *Provider) Logger() otellog.Logger {
	return p.logger
}

// Shutdown flushes and shuts down all providers.
func (p *Provider) Shutdown(ctx context.Context) {
	if err := p.tp.Shutdown(ctx); err != nil {
		debug.Log("trace provider shutdown error: %v", err)
	}
	if err := p.mp.Shutdown(ctx); err != nil {
		debug.Log("metric provider shutdown error: %v", err)
	}
	if err := p.lp.Shutdown(ctx); err != nil {
		debug.Log("log provider shutdown error: %v", err)
	}
}

func newResource(ctx context.Context, cfg config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			attribute.String("service.version", cfg.Version),
		),
	)
}

func newTracerProvider(ctx context.Context, cfg config.Config, po providerOptions, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	opts := traceExporterOpts(cfg, po)
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(newFixedIDGenerator()),
	), nil
}

func newMeterProvider(ctx context.Context, cfg config.Config, po providerOptions, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	opts := metricExporterOpts(cfg, po)
	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Use a PeriodicReader with a short interval — Shutdown() flushes remaining
	reader := sdkmetric.NewPeriodicReader(exp,
		sdkmetric.WithInterval(time.Second),
	)

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	), nil
}

func newLoggerProvider(ctx context.Context, cfg config.Config, po providerOptions, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	opts := logExporterOpts(cfg, po)
	exp, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
		sdklog.WithResource(res),
	), nil
}

// Exporter option builders — apply plugin overrides, let SDK handle standard env vars.

func traceExporterOpts(cfg config.Config, po providerOptions) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
	}
	if headers := resolveHeaders(po, cfg); len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	return opts
}

func metricExporterOpts(cfg config.Config, po providerOptions) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
	}
	if headers := resolveHeaders(po, cfg); len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	// Temporality: plugin override > standard env var > SDK default
	if pref := metricTemporality(cfg); pref != nil {
		opts = append(opts, otlpmetrichttp.WithTemporalitySelector(pref))
	}
	return opts
}

func logExporterOpts(cfg config.Config, po providerOptions) []otlploghttp.Option {
	opts := []otlploghttp.Option{
		otlploghttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlploghttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlploghttp.WithInsecure())
		}
	}
	if headers := resolveHeaders(po, cfg); len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}
	return opts
}

func resolveHeaders(po providerOptions, cfg config.Config) map[string]string {
	if len(po.headers) > 0 {
		debug.Log("resolveHeaders: using %d pre-loaded headers", len(po.headers))
		return po.headers
	}
	pluginHeaders := cfg.PluginHeaders()
	if len(pluginHeaders) > 0 {
		debug.Log("resolveHeaders: using %d plugin headers", len(pluginHeaders))
		return pluginHeaders
	}
	debug.Log("resolveHeaders: no headers available")
	return nil
}

func metricTemporality(cfg config.Config) sdkmetric.TemporalitySelector {
	pref := os.Getenv("OTEL_PLUGIN_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE")
	if pref == "" {
		pref = os.Getenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE")
	}
	if pref == "" {
		return nil // let SDK default
	}

	switch strings.ToLower(pref) {
	case "cumulative":
		return sdkmetric.DefaultTemporalitySelector
	case "delta":
		return func(sdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.DeltaTemporality
		}
	default:
		return nil
	}
}
