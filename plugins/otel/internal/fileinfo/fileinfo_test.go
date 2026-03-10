package fileinfo

import (
	"testing"
)

func TestFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
		wantExt  string
		wantLang string
	}{
		{"Go file", "/home/user/project/main.go", "/home/user/project/main.go", ".go", "Go"},
		{"Python file", "/tmp/script.py", "/tmp/script.py", ".py", "Python"},
		{"TypeScript", "/home/user/app.ts", "/home/user/app.ts", ".ts", "TypeScript"},
		{"TSX", "/home/user/App.tsx", "/home/user/App.tsx", ".tsx", "TSX"},
		{"Terraform", "/home/user/main.tf", "/home/user/main.tf", ".tf", "HCL"},
		{"Makefile", "/home/user/project/Makefile", "/home/user/project/Makefile", "Makefile", "Makefile"},
		{"Dockerfile", "/home/user/project/Dockerfile", "/home/user/project/Dockerfile", "Dockerfile", "Dockerfile"},
		{"YAML yml", "/home/user/config.yml", "/home/user/config.yml", ".yml", "YAML"},
		{"YAML yaml", "/home/user/config.yaml", "/home/user/config.yaml", ".yaml", "YAML"},
		{"JSON", "/home/user/data.json", "/home/user/data.json", ".json", "JSON"},
		{"Markdown", "/home/user/README.md", "/home/user/README.md", ".md", "Markdown"},
		{"No extension", "/home/user/project/somefile", "/home/user/project/somefile", "somefile", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := FromPath(tt.path)
			if info.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", info.Path, tt.wantPath)
			}
			if info.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", info.Extension, tt.wantExt)
			}
			if info.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", info.Language, tt.wantLang)
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
