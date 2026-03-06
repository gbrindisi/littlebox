package config

import (
	"os"
	"slices"
	"testing"
)

func TestResolveEnvPassthrough(t *testing.T) {
	// Save original environment
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range origEnv {
			parts := splitEnv(e)
			if len(parts) == 2 {
				_ = os.Setenv(parts[0], parts[1])
			}
		}
	}()

	// Set up test environment
	os.Clearenv()
	_ = os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test123")
	_ = os.Setenv("CLAUDE_CODE_FOO", "foo_value")
	_ = os.Setenv("CLAUDE_CODE_BAR", "bar_value")
	_ = os.Setenv("OPENAI_API_KEY", "sk-openai-test")
	_ = os.Setenv("OTHER_VAR", "other_value")

	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{
			name:     "exact match single",
			patterns: []string{"ANTHROPIC_API_KEY"},
			want:     []string{"ANTHROPIC_API_KEY=sk-ant-test123"},
		},
		{
			name:     "exact match multiple",
			patterns: []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"},
			want: []string{
				"ANTHROPIC_API_KEY=sk-ant-test123",
				"OPENAI_API_KEY=sk-openai-test",
			},
		},
		{
			name:     "glob pattern matches multiple",
			patterns: []string{"CLAUDE_CODE_*"},
			want: []string{
				"CLAUDE_CODE_BAR=bar_value",
				"CLAUDE_CODE_FOO=foo_value",
			},
		},
		{
			name:     "mixed exact and glob",
			patterns: []string{"ANTHROPIC_API_KEY", "CLAUDE_CODE_*"},
			want: []string{
				"ANTHROPIC_API_KEY=sk-ant-test123",
				"CLAUDE_CODE_BAR=bar_value",
				"CLAUDE_CODE_FOO=foo_value",
			},
		},
		{
			name:     "unset variable skipped",
			patterns: []string{"NONEXISTENT_VAR"},
			want:     nil,
		},
		{
			name:     "unset variable with set variable",
			patterns: []string{"NONEXISTENT_VAR", "ANTHROPIC_API_KEY"},
			want:     []string{"ANTHROPIC_API_KEY=sk-ant-test123"},
		},
		{
			name:     "glob no matches",
			patterns: []string{"NONEXISTENT_*"},
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
			name:     "glob with question mark",
			patterns: []string{"CLAUDE_CODE_???"},
			want: []string{
				"CLAUDE_CODE_BAR=bar_value",
				"CLAUDE_CODE_FOO=foo_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveEnvPassthrough(tt.patterns)

			// Sort both for comparison (env order is not guaranteed)
			slices.Sort(got)
			slices.Sort(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("ResolveEnvPassthrough() returned %d items, want %d\ngot: %v\nwant: %v",
					len(got), len(tt.want), got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ResolveEnvPassthrough()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveEnvPassthrough_ValueWithEquals(t *testing.T) {
	// Save and restore environment
	origVal, hadVal := os.LookupEnv("TEST_VAR_EQUALS")
	defer func() {
		if hadVal {
			_ = os.Setenv("TEST_VAR_EQUALS", origVal)
		} else {
			_ = os.Unsetenv("TEST_VAR_EQUALS")
		}
	}()

	// Test that values containing '=' are handled correctly
	_ = os.Setenv("TEST_VAR_EQUALS", "value=with=equals")
	got := ResolveEnvPassthrough([]string{"TEST_VAR_EQUALS"})

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}

	expected := "TEST_VAR_EQUALS=value=with=equals"
	if got[0] != expected {
		t.Errorf("got %q, want %q", got[0], expected)
	}
}

func TestResolveEnvPassthrough_EmptyValue(t *testing.T) {
	// Save and restore environment
	origVal, hadVal := os.LookupEnv("TEST_EMPTY_VAR")
	defer func() {
		if hadVal {
			_ = os.Setenv("TEST_EMPTY_VAR", origVal)
		} else {
			_ = os.Unsetenv("TEST_EMPTY_VAR")
		}
	}()

	// Test that empty values are passed through
	_ = os.Setenv("TEST_EMPTY_VAR", "")
	got := ResolveEnvPassthrough([]string{"TEST_EMPTY_VAR"})

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}

	expected := "TEST_EMPTY_VAR="
	if got[0] != expected {
		t.Errorf("got %q, want %q", got[0], expected)
	}
}

// splitEnv splits a KEY=VALUE string into [KEY, VALUE]
func splitEnv(e string) []string {
	for i := 0; i < len(e); i++ {
		if e[i] == '=' {
			return []string{e[:i], e[i+1:]}
		}
	}
	return []string{e}
}
