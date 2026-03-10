package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Context holds git metadata for the current working directory.
type Context struct {
	Branch    string
	RemoteURL string
	RepoName  string
	RepoOwner string
	HeadSHA   string
}

// GetContext reads git metadata from the given directory.
// Returns a zero-value Context (no error) if not a git repository.
func GetContext(dir string) Context {
	if !isGitRepo(dir) {
		return Context{}
	}

	ctx := Context{
		Branch:  readBranch(dir),
		HeadSHA: readHeadSHA(dir),
	}

	ctx.RemoteURL = readRemoteURL(dir)
	if ctx.RemoteURL != "" {
		ctx.RepoOwner, ctx.RepoName = ParseRemoteURL(ctx.RemoteURL)
	}

	return ctx
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// readBranch reads the current branch from .git/HEAD directly.
func readBranch(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(string(data))
	if strings.HasPrefix(ref, "ref: refs/heads/") {
		return strings.TrimPrefix(ref, "ref: refs/heads/")
	}
	// Detached HEAD — return the short SHA
	if len(ref) >= 8 {
		return ref[:8]
	}
	return ref
}

// readHeadSHA reads the HEAD commit SHA.
func readHeadSHA(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(string(data))

	// Direct SHA (detached HEAD)
	if !strings.HasPrefix(ref, "ref: ") {
		return ref
	}

	// Follow the ref to get the SHA
	refPath := strings.TrimPrefix(ref, "ref: ")
	sha, err := os.ReadFile(filepath.Join(dir, ".git", refPath))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(sha))
}

// readRemoteURL shells out to git for the origin remote URL.
func readRemoteURL(dir string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// ParseRemoteURL extracts owner and repo name from a git remote URL.
// Handles SSH (git@github.com:owner/repo.git) and
// HTTPS (https://github.com/owner/repo.git) formats.
func ParseRemoteURL(url string) (owner, repo string) {
	// Normalize: remove trailing .git
	url = strings.TrimSuffix(url, ".git")

	// SSH format: git@host:owner/repo
	if strings.Contains(url, ":") && strings.Contains(url, "@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			return splitOwnerRepo(parts[1])
		}
	}

	// HTTPS format: https://host/owner/repo
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Split by / and take last two segments
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}

	return "", ""
}

func splitOwnerRepo(path string) (string, string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", path
}
