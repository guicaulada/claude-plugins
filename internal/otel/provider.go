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
	tp              *sdktrace.TracerProvider
	mp              *sdkmetric.MeterProvider
	lp              *sdklog.LoggerProvider
	tracer          trace.Tracer
	logger          otellog.Logger
	metricBaseAttrs []attribute.KeyValue // cardinality-controlled attrs for all metrics
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

	// Apply plugin-specific protocol override via env var
	if protocol := cfg.PluginProtocol(); protocol != "" {
		_ = os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", protocol)
		debug.Log("set OTEL_EXPORTER_OTLP_PROTOCOL=%s from plugin override", protocol)
	}

	// Resolve headers once for all exporters
	headers := resolveHeaders(po, cfg)
	if len(headers) > 0 && os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") == "" {
		var pairs []string
		for k, v := range headers {
			pairs = append(pairs, k+"="+v)
		}
		_ = os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", strings.Join(pairs, ","))
		debug.Log("set OTEL_EXPORTER_OTLP_HEADERS (%d pairs)", len(headers))
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tp, err := newTracerProvider(ctx, cfg, headers, res)
	if err != nil {
		return nil, err
	}

	mp, err := newMeterProvider(ctx, cfg, headers, res)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}

	lp, err := newLoggerProvider(ctx, cfg, headers, res)
	if err != nil {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
		return nil, err
	}

	// Build cardinality-controlled metric base attrs
	var metricBaseAttrs []attribute.KeyValue
	if cfg.IncludeVersion {
		metricBaseAttrs = append(metricBaseAttrs, attribute.String("app.version", cfg.Version))
	}

	return &Provider{
		tp:              tp,
		mp:              mp,
		lp:              lp,
		tracer:          tp.Tracer(TracerName),
		logger:          lp.Logger(LoggerName),
		metricBaseAttrs: metricBaseAttrs,
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
		resource.WithFromEnv(), // reads OTEL_RESOURCE_ATTRIBUTES
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			attribute.String("service.version", cfg.Version),
		),
	)
}

func newTracerProvider(ctx context.Context, cfg config.Config, headers map[string]string, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	opts := traceExporterOpts(cfg, headers)
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp,
			sdktrace.WithBatchTimeout(100*time.Millisecond),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(newFixedIDGenerator()),
	), nil
}

func newMeterProvider(ctx context.Context, cfg config.Config, headers map[string]string, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	opts := metricExporterOpts(cfg, headers)
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

func newLoggerProvider(ctx context.Context, cfg config.Config, headers map[string]string, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	opts := logExporterOpts(cfg, headers)
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

func traceExporterOpts(cfg config.Config, headers map[string]string) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	return opts
}

func metricExporterOpts(cfg config.Config, headers map[string]string) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	// Temporality: plugin override > standard env var > SDK default
	if pref := metricTemporality(cfg); pref != nil {
		opts = append(opts, otlpmetrichttp.WithTemporalitySelector(pref))
	}
	return opts
}

func logExporterOpts(cfg config.Config, headers map[string]string) []otlploghttp.Option {
	opts := []otlploghttp.Option{
		otlploghttp.WithTimeout(2 * time.Second),
	}
	if endpoint := cfg.PluginEndpoint(); endpoint != "" {
		opts = append(opts, otlploghttp.WithEndpoint(endpoint))
		if cfg.PluginInsecure() {
			opts = append(opts, otlploghttp.WithInsecure())
		}
	}
	if len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}
	return opts
}

func resolveHeaders(po providerOptions, cfg config.Config) map[string]string {
	if len(po.headers) > 0 {
		return po.headers
	}
	return cfg.PluginHeaders()
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
