package config

import (
	"strings"
	"testing"
)

func TestImageScopeDefault(t *testing.T) {
	cfg := &Config{}
	if err := ApplyDefaults(cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Agent.ImageScope != ImageScopeWorkspace {
		t.Errorf("default image_scope = %q, want %q", cfg.Agent.ImageScope, ImageScopeWorkspace)
	}
}

func TestImageScopeValidation(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		scope   string
		wantErr bool
	}{{"", false}, {"workspace", false}, {"shared", false}, {"global", true}, {"Shared", true}} {
		cfg := &Config{Agent: AgentConfig{Command: []string{"x"}, ImageScope: tc.scope}, Workspace: WorkspaceConfig{Path: dir}}
		err := Validate(cfg)
		if (err != nil) != tc.wantErr {
			t.Errorf("scope %q: err=%v, wantErr=%v", tc.scope, err, tc.wantErr)
		}
		if err != nil && !strings.Contains(err.Error(), "agent.image_scope") {
			t.Errorf("scope %q: error should mention agent.image_scope: %v", tc.scope, err)
		}
	}
}

func TestImageScopeKey(t *testing.T) {
	cfg := &Config{Workspace: WorkspaceConfig{Path: "/a"}}
	cfg.Agent.ImageScope = ImageScopeWorkspace
	if got := cfg.ImageScopeKey(); got != "/a" {
		t.Errorf("workspace scope key = %q, want /a", got)
	}
	cfg.Agent.ImageScope = ImageScopeShared
	if got := cfg.ImageScopeKey(); got != "" {
		t.Errorf("shared scope key = %q, want empty", got)
	}
}

func TestImageScopeYAML(t *testing.T) {
	cfg, err := Parse([]byte("agent:\n  command: [x]\n  image_scope: shared\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Agent.ImageScope != ImageScopeShared {
		t.Errorf("parsed image_scope = %q", cfg.Agent.ImageScope)
	}
}
