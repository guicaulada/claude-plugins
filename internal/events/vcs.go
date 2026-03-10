package events

import (
	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	gitpkg "github.com/guicaulada/claude-code-otel-plugin/internal/git"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

// vcsAttributes reads current git context from cwd and returns OTel
// span attributes. Also tracks unique branches and repos in state
// for session-level aggregation.
func vcsAttributes(cwd, sessionID string, store *state.Store) []attribute.KeyValue {
	gitCtx := gitpkg.GetContext(cwd)
	if gitCtx.Branch == "" && gitCtx.RemoteURL == "" {
		return nil
	}

	var attrs []attribute.KeyValue

	if gitCtx.Branch != "" {
		attrs = append(attrs, attribute.String("vcs.ref.head.name", gitCtx.Branch))
		trackUnique(store, sessionID, "branch", gitCtx.Branch)
	}
	if gitCtx.RepoName != "" {
		attrs = append(attrs, attribute.String("vcs.repository.name", gitCtx.RepoName))
	}
	if gitCtx.RepoOwner != "" {
		attrs = append(attrs, attribute.String("vcs.repository.owner", gitCtx.RepoOwner))
	}
	if gitCtx.RemoteURL != "" {
		attrs = append(attrs, attribute.String("vcs.repository.url.full", gitCtx.RemoteURL))
		trackUnique(store, sessionID, "repo", gitCtx.RemoteURL)
	}

	return attrs
}

// trackUnique increments a counter only if this value hasn't been seen before.
// Uses the cache table to track seen values.
func trackUnique(store *state.Store, sessionID, category, value string) {
	key := "seen_" + category + ":" + value
	if existing, _ := store.GetCache(key); existing != "" {
		return
	}
	if err := store.SetCache(key, "1"); err != nil {
		debug.Log("failed to track unique %s: %v", category, err)
		return
	}
	_ = store.IncrementCounter(sessionID, category+"_count")
}
