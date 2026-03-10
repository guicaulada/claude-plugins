package otel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
)

// mockCollector captures OTLP export requests.
type mockCollector struct {
	mu       sync.Mutex
	requests [][]byte
	server   *httptest.Server
}

func newMockCollector() *mockCollector {
	mc := &mockCollector{}
	mc.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mc.mu.Lock()
		mc.requests = append(mc.requests, body)
		mc.mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	return mc
}

func (mc *mockCollector) requestCount() int {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return len(mc.requests)
}

func (mc *mockCollector) close() {
	mc.server.Close()
}

func TestProviderExportsSpan(t *testing.T) {
	mc := newMockCollector()
	defer mc.close()

	// Strip scheme for WithEndpoint
	endpoint := mc.server.Listener.Addr().String()

	cfg := config.Config{
		Enabled: true,
		Version: "test",
	}

	ctx := context.Background()

	// Set endpoint via env — the SDK reads it
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+endpoint)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	provider, err := NewProvider(ctx, cfg,
		WithHeaders(map[string]string{"Authorization": "Bearer test"}),
	)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}

	// Create a span
	builder := NewSpanBuilder(provider.Tracer())
	builder.CreateSpan(ctx, "test-span", time.Now(), time.Now().Add(time.Second),
		[]attribute.KeyValue{attribute.String("test", "value")},
	)

	// Shutdown flushes
	provider.Shutdown(ctx)

	// Verify requests were made
	if mc.requestCount() == 0 {
		t.Error("expected at least one export request to mock collector")
	}
}

func TestProviderEmitsLog(t *testing.T) {
	mc := newMockCollector()
	defer mc.close()

	endpoint := mc.server.Listener.Addr().String()

	cfg := config.Config{
		Enabled: true,
		Version: "test",
	}

	ctx := context.Background()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+endpoint)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	provider, err := NewProvider(ctx, cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}

	provider.EmitEvent("test.event", "", "", map[string]string{
		"key": "value",
	})

	provider.Shutdown(ctx)

	if mc.requestCount() == 0 {
		t.Error("expected at least one log export request")
	}
}

func TestProviderRecordsMetric(t *testing.T) {
	mc := newMockCollector()
	defer mc.close()

	endpoint := mc.server.Listener.Addr().String()

	cfg := config.Config{
		Enabled:        true,
		Version:        "test",
		IncludeVersion: true,
	}

	ctx := context.Background()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+endpoint)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	provider, err := NewProvider(ctx, cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}

	provider.CounterAdd(ctx, "test.counter", 1,
		attribute.String("dim", "value"),
	)

	provider.Shutdown(ctx)

	if mc.requestCount() == 0 {
		t.Error("expected at least one metric export request")
	}
}

func TestWithHeaders(t *testing.T) {
	headers := map[string]string{"X-Test": "value"}
	var po providerOptions
	WithHeaders(headers)(&po)

	if po.headers["X-Test"] != "value" {
		t.Errorf("headers = %v, want X-Test=value", po.headers)
	}
}

func TestResolveHeaders(t *testing.T) {
	po := providerOptions{
		headers: map[string]string{"X-Pre": "loaded"},
	}
	cfg := config.Config{}

	got := resolveHeaders(po, cfg)
	if got["X-Pre"] != "loaded" {
		t.Errorf("expected pre-loaded headers, got %v", got)
	}
}

func TestResolveHeadersEmpty(t *testing.T) {
	po := providerOptions{}
	cfg := config.Config{}

	got := resolveHeaders(po, cfg)
	if len(got) != 0 {
		t.Errorf("expected empty headers, got %v", got)
	}
}

// Verify JSON response doesn't break the mock
func TestMockCollectorJSON(t *testing.T) {
	mc := newMockCollector()
	defer mc.close()

	resp, err := http.Post(mc.server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("response should be valid JSON: %v", err)
	}
}
