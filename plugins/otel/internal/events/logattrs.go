package events

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"

	gitpkg "github.com/guicaulada/claude-plugins/plugins/otel/internal/git"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/payload"
	"github.com/guicaulada/claude-plugins/plugins/otel/internal/state"
)

// commonLogAttrs returns log attributes that every event should include.
func commonLogAttrs(env payload.Envelope) []log.KeyValue {
	attrs := []log.KeyValue{
		log.String("claude_code.session.id", env.SessionID),
		log.String("claude_code.session.cwd", env.Cwd),
	}
	if env.PermissionMode != "" {
		attrs = append(attrs, log.String("claude_code.permission_mode", env.PermissionMode))
	}

	gitCtx := gitpkg.GetContext(env.Cwd)
	if gitCtx.Branch != "" {
		attrs = append(attrs, log.String("vcs.ref.head.name", gitCtx.Branch))
	}
	if gitCtx.RepoName != "" {
		attrs = append(attrs, log.String("vcs.repository.name", gitCtx.RepoName))
	}
	if gitCtx.RepoOwner != "" {
		attrs = append(attrs, log.String("vcs.repository.owner", gitCtx.RepoOwner))
	}
	if gitCtx.RemoteURL != "" {
		attrs = append(attrs, log.String("vcs.repository.url.full", gitCtx.RemoteURL))
	}

	return attrs
}

// commonLogAttrsFromSession returns log attributes from stored session state.
func commonLogAttrsFromSession(env payload.Envelope, sess state.Session) []log.KeyValue {
	attrs := []log.KeyValue{
		log.String("claude_code.session.id", env.SessionID),
		log.String("claude_code.session.cwd", sess.Cwd),
	}
	if sess.GitBranch != "" {
		attrs = append(attrs, log.String("vcs.ref.head.name", sess.GitBranch))
	}
	if sess.GitRepoName != "" {
		attrs = append(attrs, log.String("vcs.repository.name", sess.GitRepoName))
	}
	if sess.GitRepoOwner != "" {
		attrs = append(attrs, log.String("vcs.repository.owner", sess.GitRepoOwner))
	}
	if sess.GitRemoteURL != "" {
		attrs = append(attrs, log.String("vcs.repository.url.full", sess.GitRemoteURL))
	}
	return attrs
}

// vcsMetricAttrs returns VCS attributes suitable for metrics (no URL — high cardinality).
func vcsMetricAttrs(cwd string) []attribute.KeyValue {
	gitCtx := gitpkg.GetContext(cwd)
	var attrs []attribute.KeyValue
	if gitCtx.Branch != "" {
		attrs = append(attrs, attribute.String("vcs.ref.head.name", gitCtx.Branch))
	}
	if gitCtx.RepoName != "" {
		attrs = append(attrs, attribute.String("vcs.repository.name", gitCtx.RepoName))
	}
	if gitCtx.RepoOwner != "" {
		attrs = append(attrs, attribute.String("vcs.repository.owner", gitCtx.RepoOwner))
	}
	return attrs
}

// vcsMetricAttrsFromSession returns VCS metric attributes from stored session state.
func vcsMetricAttrsFromSession(sess state.Session) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if sess.GitBranch != "" {
		attrs = append(attrs, attribute.String("vcs.ref.head.name", sess.GitBranch))
	}
	if sess.GitRepoName != "" {
		attrs = append(attrs, attribute.String("vcs.repository.name", sess.GitRepoName))
	}
	if sess.GitRepoOwner != "" {
		attrs = append(attrs, attribute.String("vcs.repository.owner", sess.GitRepoOwner))
	}
	return attrs
}
