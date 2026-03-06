package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestGetRedactedFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files and directories
	files := []string{
		".env",
		"config.yaml",
		"server.pem",
		"client.pem",
		"README.md",
		"secrets/api_key.txt",
		"secrets/database.conf",
		"src/main.go",
		"src/utils.go",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{
			name:     "exact filename matches .env",
			patterns: []string{".env"},
			want:     []string{".env"},
		},
		{
			name:     "glob *.pem matches all pem files",
			patterns: []string{"*.pem"},
			want:     []string{"client.pem", "server.pem"},
		},
		{
			name:     "directory pattern secrets/*",
			patterns: []string{"secrets/*"},
			want:     []string{"secrets/api_key.txt", "secrets/database.conf"},
		},
		{
			name:     "multiple patterns",
			patterns: []string{".env", "*.pem"},
			want:     []string{".env", "client.pem", "server.pem"},
		},
		{
			name:     "no matches",
			patterns: []string{"*.xyz"},
			want:     nil,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			want:     nil,
		},
		{
			name:     "nil patterns",
			patterns: nil,
			want:     nil,
		},
		{
			name:     "question mark glob",
			patterns: []string{"*.??"},
			want:     []string{"src/main.go", "src/utils.go", "README.md"},
		},
		{
			name:     "match specific extension in directory",
			patterns: []string{"*.go"},
			want:     []string{"src/main.go", "src/utils.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRedactedFiles(tmpDir, tt.patterns)
			if err != nil {
				t.Fatalf("GetRedactedFiles() error = %v", err)
			}

			// Convert to relative paths for comparison
			gotRel := make([]string, len(got))
			for i, p := range got {
				rel, _ := filepath.Rel(tmpDir, p)
				gotRel[i] = rel
			}

			// Sort for comparison
			slices.Sort(gotRel)
			slices.Sort(tt.want)

			if len(gotRel) != len(tt.want) {
				t.Errorf("GetRedactedFiles() returned %d items, want %d\ngot: %v\nwant: %v",
					len(gotRel), len(tt.want), gotRel, tt.want)
				return
			}

			for i := range gotRel {
				if gotRel[i] != tt.want[i] {
					t.Errorf("GetRedactedFiles()[%d] = %q, want %q", i, gotRel[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetRedactedFiles_ReturnsAbsolutePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	got, err := GetRedactedFiles(tmpDir, []string{".env"})
	if err != nil {
		t.Fatalf("GetRedactedFiles() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}

	if !filepath.IsAbs(got[0]) {
		t.Errorf("expected absolute path, got %q", got[0])
	}

	if got[0] != testFile {
		t.Errorf("got %q, want %q", got[0], testFile)
	}
}

func TestGetRedactedFiles_CaseSensitive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different cases
	files := []string{".env", ".ENV", ".Env"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}

	got, err := GetRedactedFiles(tmpDir, []string{".env"})
	if err != nil {
		t.Fatalf("GetRedactedFiles() error = %v", err)
	}

	// Should only match exactly ".env" (case-sensitive)
	if len(got) != 1 {
		t.Errorf("expected 1 match (case-sensitive), got %d: %v", len(got), got)
	}
}

func TestGetRedactedFiles_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deeply nested .env files
	paths := []string{
		".env",
		"config/.env",
		"deeply/nested/dir/.env",
	}

	for _, p := range paths {
		fullPath := filepath.Join(tmpDir, p)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	got, err := GetRedactedFiles(tmpDir, []string{".env"})
	if err != nil {
		t.Fatalf("GetRedactedFiles() error = %v", err)
	}

	// Should match all .env files by basename
	if len(got) != 3 {
		t.Errorf("expected 3 matches, got %d: %v", len(got), got)
	}
}

// TestGetRedactedFiles_CommonPatterns tests common sensitive file patterns
// as specified in task agent-box-ymu.3 acceptance criteria.
func TestGetRedactedFiles_CommonPatterns(t *testing.T) {
	dir := t.TempDir()

	// Create test files matching the spec
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET"), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "key.pem"), []byte("KEY"), 0644); err != nil {
		t.Fatalf("failed to create key.pem: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "secrets"), 0755); err != nil {
		t.Fatalf("failed to create secrets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secrets", "api.key"), []byte("KEY"), 0644); err != nil {
		t.Fatalf("failed to create secrets/api.key: %v", err)
	}

	tests := []struct {
		patterns []string
		want     int
	}{
		{[]string{".env"}, 1},
		{[]string{"*.pem"}, 1},
		{[]string{".env", "*.pem"}, 2},
		{[]string{"secrets/*"}, 1},
	}

	for _, tt := range tests {
		got, err := GetRedactedFiles(dir, tt.patterns)
		if err != nil {
			t.Errorf("GetRedactedFiles(%v) error = %v", tt.patterns, err)
			continue
		}
		if len(got) != tt.want {
			t.Errorf("GetRedactedFiles(%v) = %d files, want %d", tt.patterns, len(got), tt.want)
		}
	}
}

// TestGetRedactedFiles_NestedEnvFile tests secrets/.env pattern specifically.
func TestGetRedactedFiles_NestedEnvFile(t *testing.T) {
	dir := t.TempDir()

	// Create nested .env file
	if err := os.MkdirAll(filepath.Join(dir, "secrets"), 0755); err != nil {
		t.Fatalf("failed to create secrets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secrets", ".env"), []byte("SECRET"), 0644); err != nil {
		t.Fatalf("failed to create secrets/.env: %v", err)
	}
	// Create other files that should NOT match
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Test that .env pattern matches nested .env file
	got, err := GetRedactedFiles(dir, []string{".env"})
	if err != nil {
		t.Fatalf("GetRedactedFiles() error = %v", err)
	}

	if len(got) != 1 {
		t.Errorf("expected 1 match for secrets/.env, got %d: %v", len(got), got)
	}

	// Test that main.go is NOT in the redacted list (other files are readable)
	for _, p := range got {
		if filepath.Base(p) == "main.go" {
			t.Error("main.go should NOT be in redacted files")
		}
	}
}

func TestGetRedactedFiles_NonExistentWorkspace(t *testing.T) {
	got, err := GetRedactedFiles("/nonexistent/path", []string{".env"})
	// Non-existent workspace silently returns empty results (errors ignored during walk)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty results, got %v", got)
	}
}
