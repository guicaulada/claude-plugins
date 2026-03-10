package git

import (
	"path/filepath"
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "HTTPS with .git",
			url:       "https://github.com/guicaulada/claude-code-otel-plugin.git",
			wantOwner: "guicaulada",
			wantRepo:  "claude-code-otel-plugin",
		},
		{
			name:      "HTTPS without .git",
			url:       "https://github.com/guicaulada/claude-code-otel-plugin",
			wantOwner: "guicaulada",
			wantRepo:  "claude-code-otel-plugin",
		},
		{
			name:      "SSH with .git",
			url:       "git@github.com:guicaulada/claude-code-otel-plugin.git",
			wantOwner: "guicaulada",
			wantRepo:  "claude-code-otel-plugin",
		},
		{
			name:      "SSH without .git",
			url:       "git@github.com:guicaulada/claude-code-otel-plugin",
			wantOwner: "guicaulada",
			wantRepo:  "claude-code-otel-plugin",
		},
		{
			name:      "GitLab HTTPS",
			url:       "https://gitlab.com/org/subgroup/repo.git",
			wantOwner: "subgroup",
			wantRepo:  "repo",
		},
		{
			name:      "empty",
			url:       "",
			wantOwner: "",
			wantRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo := ParseRemoteURL(tt.url)
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

func TestGetContext(t *testing.T) {
	// Test against the current repo using absolute path
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	ctx := GetContext(repoRoot)
	if ctx.Branch == "" {
		t.Error("expected non-empty branch")
	}
	if ctx.HeadSHA == "" {
		t.Error("expected non-empty HEAD SHA")
	}
	// Remote URL may be empty if no remote is configured
	t.Logf("branch=%s remote=%s owner=%s repo=%s sha=%s",
		ctx.Branch, ctx.RemoteURL, ctx.RepoOwner, ctx.RepoName, ctx.HeadSHA)
}

func TestGetContextNonGitDir(t *testing.T) {
	ctx := GetContext("/tmp")
	if ctx.Branch != "" {
		t.Errorf("expected empty branch for non-git dir, got %q", ctx.Branch)
	}
	if ctx.RemoteURL != "" {
		t.Errorf("expected empty remote URL for non-git dir, got %q", ctx.RemoteURL)
	}
}
