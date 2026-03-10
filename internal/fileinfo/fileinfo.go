package fileinfo

import (
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// Info holds metadata about a file path.
type Info struct {
	Path      string // Absolute file path as-is
	Extension string // Raw extension (e.g., ".go") or filename (e.g., "Makefile")
	Language  string // Language name via go-enry (e.g., "Go", "Python")
}

// FromPath extracts file metadata from an absolute file path.
func FromPath(absPath string) Info {
	info := Info{
		Path: absPath,
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

// DiffLines computes lines added and removed between old and new content
// by comparing line-by-line. Lines present in new but not old are added;
// lines present in old but not new are removed.
func DiffLines(oldContent, newContent string) (added, removed int) {
	if oldContent == "" && newContent == "" {
		return 0, 0
	}
	if oldContent == "" {
		return CountLines(newContent), 0
	}
	if newContent == "" {
		return 0, CountLines(oldContent)
	}

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Build a frequency map of old lines
	oldFreq := make(map[string]int)
	for _, line := range oldLines {
		oldFreq[line]++
	}

	// Count lines in new that aren't in old
	for _, line := range newLines {
		if oldFreq[line] > 0 {
			oldFreq[line]--
		} else {
			added++
		}
	}

	// Remaining old lines that weren't matched
	for _, count := range oldFreq {
		removed += count
	}

	return added, removed
}
