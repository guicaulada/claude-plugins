package fileinfo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// Info holds metadata about a file path.
type Info struct {
	Path      string // Relative to $HOME using ~, or absolute
	Extension string // Raw extension (e.g., ".go") or filename (e.g., "Makefile")
	Language  string // Language name via go-enry (e.g., "Go", "Python")
}

// FromPath extracts file metadata from an absolute file path.
func FromPath(absPath string) Info {
	info := Info{
		Path: relativizePath(absPath),
	}

	base := filepath.Base(absPath)
	ext := filepath.Ext(base)

	if ext != "" {
		info.Extension = ext
	} else {
		// No extension — use the filename (e.g., Makefile, Dockerfile)
		info.Extension = base
	}

	// Language detection: check overrides first for ambiguous extensions,
	// then fall back to go-enry
	lang := languageOverrides[info.Extension]
	if lang == "" {
		lang, _ = enry.GetLanguageByFilename(base)
	}
	if lang == "" {
		lang, _ = enry.GetLanguageByExtension(base)
	}
	info.Language = lang

	return info
}

// relativizePath converts an absolute path to ~ relative if under $HOME.
func relativizePath(absPath string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return absPath
	}

	if strings.HasPrefix(absPath, home) {
		rel, err := filepath.Rel(home, absPath)
		if err == nil {
			return "~/" + rel
		}
	}

	return absPath
}

// languageOverrides fixes go-enry's ambiguous extension results.
var languageOverrides = map[string]string{
	".md":       "Markdown",
	".markdown": "Markdown",
	".yml":      "YAML",
	".yaml":     "YAML",
}

// CountLines returns the number of lines in a string.
func CountLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// DiffLines computes lines added and removed between old and new content.
func DiffLines(oldContent, newContent string) (added, removed int) {
	oldLines := CountLines(oldContent)
	newLines := CountLines(newContent)

	if newLines > oldLines {
		return newLines - oldLines, 0
	}
	return 0, oldLines - newLines
}
