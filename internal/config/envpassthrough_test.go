package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnvPassthroughEntry_UnmarshalYAML(t *testing.T) {
	src := `
env_passthrough:
  - ANTHROPIC_API_KEY
  - name: OPENAI_API_KEY
    required: true
  - {name: GEMINI_API_KEY}
  - CLAUDE_CODE_*
`
	var out struct {
		EnvPassthrough []EnvPassthroughEntry `yaml:"env_passthrough"`
	}
	if err := yaml.Unmarshal([]byte(src), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []EnvPassthroughEntry{
		{Name: "ANTHROPIC_API_KEY"},
		{Name: "OPENAI_API_KEY", Required: true},
		{Name: "GEMINI_API_KEY"},
		{Name: "CLAUDE_CODE_*"},
	}
	if len(out.EnvPassthrough) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(out.EnvPassthrough), len(want), out.EnvPassthrough)
	}
	for i := range want {
		if out.EnvPassthrough[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, out.EnvPassthrough[i], want[i])
		}
	}
}

func TestEnvPassthroughEntry_UnmarshalYAML_Invalid(t *testing.T) {
	for name, src := range map[string]string{
		"missing name": "- {required: true}",
		"sequence":     "- [FOO]",
	} {
		t.Run(name, func(t *testing.T) {
			var out []EnvPassthroughEntry
			if err := yaml.Unmarshal([]byte(src), &out); err == nil {
				t.Errorf("expected error, got %+v", out)
			}
		})
	}
}
