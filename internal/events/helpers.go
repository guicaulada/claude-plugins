package events

import (
	"context"
	"encoding/json"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	pluginotel "github.com/guicaulada/claude-code-otel-plugin/internal/otel"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

// newProviderFromState creates an OTel provider using cached headers from state.
// If no cache exists, loads headers from otelHeadersHelper and caches them.
func newProviderFromState(ctx context.Context, cfg config.Config, store *state.Store) (*pluginotel.Provider, error) {
	var opts []pluginotel.ProviderOption

	if cached, err := store.GetCache("otel_headers"); err == nil && cached != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(cached), &headers); err == nil && len(headers) > 0 {
			opts = append(opts, pluginotel.WithHeaders(headers))
			debug.Log("using %d cached OTel headers", len(headers))
		}
	} else {
		// No cache — load from otelHeadersHelper and cache for future invocations
		headers := config.LoadOTelHeaders()
		if len(headers) > 0 {
			opts = append(opts, pluginotel.WithHeaders(headers))
			headersJSON, err := json.Marshal(headers)
			if err == nil {
				_ = store.SetCache("otel_headers", string(headersJSON))
				debug.Log("loaded and cached %d OTel headers from helper", len(headers))
			}
		}
	}

	return pluginotel.NewProvider(ctx, cfg, opts...)
}
