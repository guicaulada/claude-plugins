package fileinfo

import (
	"os"
	"testing"
)

func TestFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantExt  string
		wantLang string
	}{
		{"Go file", "/home/user/project/main.go", ".go", "Go"},
		{"Python file", "/tmp/script.py", ".py", "Python"},
		{"TypeScript", "/home/user/app.ts", ".ts", "TypeScript"},
		{"TSX", "/home/user/App.tsx", ".tsx", "TSX"},
		{"Terraform", "/home/user/main.tf", ".tf", "HCL"},
		{"Makefile", "/home/user/project/Makefile", "Makefile", "Makefile"},
		{"Dockerfile", "/home/user/project/Dockerfile", "Dockerfile", "Dockerfile"},
		{"YAML yml", "/home/user/config.yml", ".yml", "YAML"},
		{"YAML yaml", "/home/user/config.yaml", ".yaml", "YAML"},
		{"JSON", "/home/user/data.json", ".json", "JSON"},
		{"Markdown", "/home/user/README.md", ".md", "Markdown"},
		{"No extension unknown", "/home/user/project/somefile", "somefile", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := FromPath(tt.path)
			if info.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", info.Extension, tt.wantExt)
			}
			if info.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", info.Language, tt.wantLang)
			}
		})
	}
}

func TestRelativizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"under home", home + "/Work/project/file.go", "~/Work/project/file.go"},
		{"outside home", "/tmp/file.go", "/tmp/file.go"},
		{"home root", home + "/file.go", "~/file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relativizePath(tt.path)
			if got != tt.want {
				t.Errorf("relativizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"one line", 1},
		{"line1\nline2", 2},
		{"line1\nline2\nline3\n", 4},
	}

	for _, tt := range tests {
		got := CountLines(tt.input)
		if got != tt.want {
			t.Errorf("CountLines(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestDiffLines(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		new         string
		wantAdded   int
		wantRemoved int
	}{
		{"add lines", "a\nb", "a\nb\nc\nd", 2, 0},
		{"remove lines", "a\nb\nc", "a", 0, 2},
		{"same", "a\nb", "a\nb", 0, 0},
		{"empty to content", "", "a\nb\nc", 3, 0},
		{"content to empty", "a\nb", "", 0, 2},
		{
			"replace lines",
			"line1\nold line\nline3",
			"line1\nnew line\nline3",
			1, 1,
		},
		{
			"mixed add and remove",
			"a\nb\nc\nd",
			"a\nx\ny\nc",
			2, 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed := DiffLines(tt.old, tt.new)
			if added != tt.wantAdded {
				t.Errorf("added = %d, want %d", added, tt.wantAdded)
			}
			if removed != tt.wantRemoved {
				t.Errorf("removed = %d, want %d", removed, tt.wantRemoved)
			}
		})
	}
}
