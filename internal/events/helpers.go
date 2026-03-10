package events

import (
	"context"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	pluginotel "github.com/guicaulada/claude-code-otel-plugin/internal/otel"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

// newProviderFromState creates an OTel provider using shared cached headers.
// The header cache is shared across all sessions via a temp file with
// timestamp-based expiry controlled by CLAUDE_CODE_OTEL_HEADERS_HELPER_DEBOUNCE_MS.
func newProviderFromState(ctx context.Context, cfg config.Config, store *state.Store) (*pluginotel.Provider, error) {
	var opts []pluginotel.ProviderOption

	if headers := config.LoadCachedHeaders(); len(headers) > 0 {
		opts = append(opts, pluginotel.WithHeaders(headers))
	}

	return pluginotel.NewProvider(ctx, cfg, opts...)
}
