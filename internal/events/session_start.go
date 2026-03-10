package events

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/guicaulada/claude-code-otel-plugin/internal/config"
	"github.com/guicaulada/claude-code-otel-plugin/internal/debug"
	gitpkg "github.com/guicaulada/claude-code-otel-plugin/internal/git"
	"github.com/guicaulada/claude-code-otel-plugin/internal/idgen"
	"github.com/guicaulada/claude-code-otel-plugin/internal/payload"
	"github.com/guicaulada/claude-code-otel-plugin/internal/state"
)

type sessionStartEvent struct {
	Source string `json:"source"`
	Model  string `json:"model"`
}

func HandleSessionStart(env payload.Envelope) error {
	var event sessionStartEvent
	if err := json.Unmarshal(env.RawEvent, &event); err != nil {
		debug.Log("failed to parse SessionStart event: %v", err)
	}

	store, err := state.Open(env.SessionID)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	// Extract git context from the working directory
	gitCtx := gitpkg.GetContext(env.Cwd)

	sess := state.Session{
		SessionID: env.SessionID,
		TraceID:   idgen.TraceID(),
		SpanID:    idgen.SpanID(),
		StartTime: time.Now().UnixNano(),
		Cwd:       env.Cwd,
		StartType: event.Source,
		GitBranch:      gitCtx.Branch,
		GitRemoteURL:   gitCtx.RemoteURL,
		GitRepoName:    gitCtx.RepoName,
		GitRepoOwner:   gitCtx.RepoOwner,
	}

	if err := store.CreateSession(sess); err != nil {
		return err
	}

	// Cache git HEAD SHA for commit detection in Bash tool handlers
	if gitCtx.HeadSHA != "" {
		_ = store.SetCache("git_head_sha", gitCtx.HeadSHA)
	}

	// Emit event and metric (newProvider handles header caching)
	ctx := context.Background()
	cfg := config.Load()
	provider, err := newProvider(ctx, cfg)
	if err != nil {
		debug.Log("session start: failed to create provider: %v", err)
	} else {
		defer provider.Shutdown(ctx)

		logAttrs := commonLogAttrs(env)
		logAttrs["claude_code.session.start_type"] = event.Source
		provider.EmitEvent("claude_code.session.start", sess.TraceID, sess.SpanID, logAttrs)

		metricAttrs := []attribute.KeyValue{
			attribute.String("claude_code.session.start_type", event.Source),
			attribute.String("claude_code.session.cwd", env.Cwd),
		}
		metricAttrs = append(metricAttrs, vcsMetricAttrs(env.Cwd)...)
		provider.CounterAdd(ctx, "claude_code.session.count", 1, metricAttrs...)
	}

	debug.Log("session start: %s (trace: %s, type: %s, cwd: %s, branch: %s, repo: %s/%s)",
		env.SessionID, sess.TraceID, event.Source, env.Cwd,
		gitCtx.Branch, gitCtx.RepoOwner, gitCtx.RepoName)
	return nil
}
