package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde with path expands to home directory",
			input:    "~/foo",
			expected: filepath.Join(home, "foo"),
		},
		{
			name:     "tilde alone expands to home directory",
			input:    "~",
			expected: home,
		},
		{
			name:     "tilde with nested path expands correctly",
			input:    "~/workspace/.claude",
			expected: filepath.Join(home, "workspace/.claude"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "tilde user path NOT expanded",
			input:    "~user/path",
			expected: "~user/path",
		},
		{
			name:     "tilde with different username NOT expanded",
			input:    "~otheruser/documents",
			expected: "~otheruser/documents",
		},
		{
			name:     "empty path unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde in middle of path NOT expanded",
			input:    "/home/~user/path",
			expected: "/home/~user/path",
		},
		{
			name:     "path with tilde after slash NOT expanded",
			input:    "foo/~/bar",
			expected: "foo/~/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandTilde(tt.input)
			if result != tt.expected {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		baseDir  string
		expected string
	}{
		{
			name:     "current directory resolves to baseDir",
			path:     ".",
			baseDir:  "/projects/myapp",
			expected: "/projects/myapp",
		},
		{
			name:     "parent directory reference resolves correctly",
			path:     "../shared",
			baseDir:  "/projects/myapp",
			expected: "/projects/shared",
		},
		{
			name:     "tilde path ignores baseDir",
			path:     "~/config",
			baseDir:  "/any",
			expected: filepath.Join(home, "config"),
		},
		{
			name:     "absolute path unchanged",
			path:     "/absolute",
			baseDir:  "/any",
			expected: "/absolute",
		},
		{
			name:     "relative path resolved against baseDir",
			path:     "src/main",
			baseDir:  "/projects/myapp",
			expected: "/projects/myapp/src/main",
		},
		{
			name:     "empty path returns empty",
			path:     "",
			baseDir:  "/any",
			expected: "",
		},
		{
			name:     "tilde alone returns home",
			path:     "~",
			baseDir:  "/any",
			expected: home,
		},
		{
			name:     "tilde user path NOT expanded but resolved",
			path:     "~user/path",
			baseDir:  "/projects",
			expected: "/projects/~user/path",
		},
		{
			name:     "nested relative path",
			path:     "../../other/project",
			baseDir:  "/a/b/c",
			expected: "/a/other/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePath(tt.path, tt.baseDir)
			if result != tt.expected {
				t.Errorf("resolvePath(%q, %q) = %q, want %q", tt.path, tt.baseDir, result, tt.expected)
			}
		})
	}
}

func TestConfigExpandPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	cfg := &Config{
		Workspace: WorkspaceConfig{
			Path: "~/workspace",
		},
		Mounts: []MountConfig{
			{Source: "~/.claude", Target: "/config"},
			{Source: "/absolute/path", Target: "/data"},
			{Source: "~user/path", Target: "/user"},
		},
	}

	cfg.ExpandPaths("/base")

	// Check workspace path expanded (tilde takes precedence, baseDir ignored)
	expectedWorkspace := filepath.Join(home, "workspace")
	if cfg.Workspace.Path != expectedWorkspace {
		t.Errorf("Workspace.Path = %q, want %q", cfg.Workspace.Path, expectedWorkspace)
	}

	// Check first mount expanded (tilde takes precedence, baseDir ignored)
	expectedMount0 := filepath.Join(home, ".claude")
	if cfg.Mounts[0].Source != expectedMount0 {
		t.Errorf("Mounts[0].Source = %q, want %q", cfg.Mounts[0].Source, expectedMount0)
	}

	// Check absolute path unchanged
	if cfg.Mounts[1].Source != "/absolute/path" {
		t.Errorf("Mounts[1].Source = %q, want %q", cfg.Mounts[1].Source, "/absolute/path")
	}

	// Check ~user path resolved relative to baseDir (since ~user is not expanded)
	expectedMount2 := "/base/~user/path"
	if cfg.Mounts[2].Source != expectedMount2 {
		t.Errorf("Mounts[2].Source = %q, want %q", cfg.Mounts[2].Source, expectedMount2)
	}
}

func TestConfigExpandPathsRelative(t *testing.T) {
	cfg := &Config{
		Workspace: WorkspaceConfig{
			Path: ".",
		},
		Mounts: []MountConfig{
			{Source: "../shared/config", Target: "/config"},
			{Source: "local/data", Target: "/data"},
		},
	}

	cfg.ExpandPaths("/projects/myapp")

	// Check workspace path resolved to baseDir
	if cfg.Workspace.Path != "/projects/myapp" {
		t.Errorf("Workspace.Path = %q, want %q", cfg.Workspace.Path, "/projects/myapp")
	}

	// Check parent directory mount resolved
	expectedMount0 := "/projects/shared/config"
	if cfg.Mounts[0].Source != expectedMount0 {
		t.Errorf("Mounts[0].Source = %q, want %q", cfg.Mounts[0].Source, expectedMount0)
	}

	// Check relative path mount resolved
	expectedMount1 := "/projects/myapp/local/data"
	if cfg.Mounts[1].Source != expectedMount1 {
		t.Errorf("Mounts[1].Source = %q, want %q", cfg.Mounts[1].Source, expectedMount1)
	}
}

func TestExpandPathExported(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Test that the exported ExpandPath function works
	result := ExpandPath("~/test")
	expected := filepath.Join(home, "test")
	if result != expected {
		t.Errorf("ExpandPath(\"~/test\") = %q, want %q", result, expected)
	}

	// Test that ~user is not expanded
	result = ExpandPath("~user/test")
	if result != "~user/test" {
		t.Errorf("ExpandPath(\"~user/test\") = %q, want %q", result, "~user/test")
	}
}
