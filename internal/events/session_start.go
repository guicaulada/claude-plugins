package events

import (
	"encoding/json"
	"time"

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
	defer store.Close()

	// Extract git context from the working directory
	gitCtx := gitpkg.GetContext(env.Cwd)

	sess := state.Session{
		SessionID:      env.SessionID,
		TraceID:        idgen.TraceID(),
		SpanID:         idgen.SpanID(),
		StartTime:      time.Now().UnixNano(),
		Cwd:            env.Cwd,
		PermissionMode: env.PermissionMode,
		StartType:      event.Source,
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

	// Cache OTel headers from otelHeadersHelper so SessionEnd
	// doesn't need to re-run the script (which may fail during shutdown)
	headers := config.LoadOTelHeaders()
	if len(headers) > 0 {
		headersJSON, err := json.Marshal(headers)
		if err == nil {
			_ = store.SetCache("otel_headers", string(headersJSON))
			debug.Log("cached %d OTel headers in state", len(headers))
		}
	}

	debug.Log("session start: %s (trace: %s, type: %s, cwd: %s, branch: %s, repo: %s/%s, permission_mode: %q)",
		env.SessionID, sess.TraceID, event.Source, env.Cwd,
		gitCtx.Branch, gitCtx.RepoOwner, gitCtx.RepoName, env.PermissionMode)
	return nil
}
