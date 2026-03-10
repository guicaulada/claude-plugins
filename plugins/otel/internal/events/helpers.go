package events

import (
	"context"

	"github.com/guicaulada/claude-plugins/plugins/otel/internal/config"
	pluginotel "github.com/guicaulada/claude-plugins/plugins/otel/internal/otel"
)

// newProvider creates an OTel provider using shared cached headers.
// The header cache is shared across all sessions via a temp file with
// timestamp-based expiry controlled by CLAUDE_CODE_OTEL_HEADERS_HELPER_DEBOUNCE_MS.
func newProvider(ctx context.Context, cfg config.Config) (*pluginotel.Provider, error) {
	var opts []pluginotel.ProviderOption

	if headers := config.LoadCachedHeaders(); len(headers) > 0 {
		opts = append(opts, pluginotel.WithHeaders(headers))
	}

	return pluginotel.NewProvider(ctx, cfg, opts...)
}
