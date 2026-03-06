package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name: "minimal config",
			yaml: `
agent:
  command: ["claude"]
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Agent.Command) != 1 || cfg.Agent.Command[0] != "claude" {
					t.Errorf("expected command [claude], got %v", cfg.Agent.Command)
				}
			},
		},
		{
			name: "custom agent",
			yaml: `
agent:
  command: ["run", "--flag"]
  env:
    - name: MY_VAR
      value: my_value
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Agent.Command) != 2 || cfg.Agent.Command[0] != "run" {
					t.Errorf("expected command [run --flag], got %v", cfg.Agent.Command)
				}
				if len(cfg.Agent.Env) != 1 || cfg.Agent.Env[0].Name != "MY_VAR" {
					t.Errorf("expected env MY_VAR, got %v", cfg.Agent.Env)
				}
			},
		},
		{
			name: "agent with build_script",
			yaml: `
agent:
  command: ["claude"]
  build_script: |
    curl -fsSL https://claude.ai/install.sh | bash
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Agent.BuildScript == "" {
					t.Error("expected build_script to be set")
				}
				if len(cfg.Agent.Command) != 1 || cfg.Agent.Command[0] != "claude" {
					t.Errorf("expected command [claude], got %v", cfg.Agent.Command)
				}
			},
		},
		{
			name: "workspace config",
			yaml: `
agent:
  command: ["test"]
workspace:
  path: /my/path
  writable: false
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Workspace.Path != "/my/path" {
					t.Errorf("expected path /my/path, got %s", cfg.Workspace.Path)
				}
				if cfg.Workspace.Writable == nil || *cfg.Workspace.Writable != false {
					t.Errorf("expected writable false, got %v", cfg.Workspace.Writable)
				}
			},
		},
		{
			name: "mounts config",
			yaml: `
agent:
  command: ["test"]
mounts:
  - source: /host/path
    target: /container/path
    readonly: true
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Mounts) != 1 {
					t.Fatalf("expected 1 mount, got %d", len(cfg.Mounts))
				}
				m := cfg.Mounts[0]
				if m.Source != "/host/path" || m.Target != "/container/path" || !m.ReadOnly {
					t.Errorf("unexpected mount: %+v", m)
				}
			},
		},
		{
			name: "network config with presets list",
			yaml: `
agent:
  command: ["test"]
network:
  presets:
    - anthropic
    - github
  allow:
    - example.com
  deny_all_else: true
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Network.Presets) != 2 {
					t.Errorf("expected 2 presets, got %d", len(cfg.Network.Presets))
				}
				if cfg.Network.Presets[0] != "anthropic" || cfg.Network.Presets[1] != "github" {
					t.Errorf("expected presets [anthropic, github], got %v", cfg.Network.Presets)
				}
				if len(cfg.Network.Allow) != 1 || cfg.Network.Allow[0] != "example.com" {
					t.Errorf("expected allow [example.com], got %v", cfg.Network.Allow)
				}
				if !cfg.Network.DenyAllElse {
					t.Error("expected deny_all_else true")
				}
			},
		},
		{
			name: "container config",
			yaml: `
agent:
  command: ["test"]
container:
  keep: true
  readonly_root: true
`,
			check: func(t *testing.T, cfg *Config) {
				if !cfg.Container.Keep {
					t.Error("expected keep true")
				}
				if !cfg.Container.ReadOnlyRoot {
					t.Error("expected readonly_root true")
				}
			},
		},
		{
			name: "container security config",
			yaml: `
agent:
  command: ["test"]
container:
  no_new_privileges: true
  seccomp_profile: /path/to/seccomp.json
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Container.NoNewPrivileges == nil || !*cfg.Container.NoNewPrivileges {
					t.Error("expected no_new_privileges true")
				}
				if cfg.Container.SeccompProfile != "/path/to/seccomp.json" {
					t.Errorf("expected seccomp_profile /path/to/seccomp.json, got %s", cfg.Container.SeccompProfile)
				}
			},
		},
		{
			name: "redact config",
			yaml: `
agent:
  command: ["test"]
redact:
  - ".env"
  - "**/*.pem"
`,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Redact) != 2 {
					t.Fatalf("expected 2 redact patterns, got %d", len(cfg.Redact))
				}
				if cfg.Redact[0] != ".env" || cfg.Redact[1] != "**/*.pem" {
					t.Errorf("unexpected redact patterns: %v", cfg.Redact)
				}
			},
		},
		{
			name:    "invalid yaml",
			yaml:    `invalid: [`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && cfg != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Run("applies workspace defaults", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{Command: []string{"test"}},
		}

		err := ApplyDefaults(cfg)
		if err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		// Workspace path defaults to "." for portability - resolved relative to Agentfile by ExpandPaths()
		if cfg.Workspace.Path != "." {
			t.Errorf("expected workspace path \".\", got %s", cfg.Workspace.Path)
		}
		if cfg.Workspace.MountPoint != DefaultWorkspaceMountPoint {
			t.Errorf("expected mount point %s, got %s", DefaultWorkspaceMountPoint, cfg.Workspace.MountPoint)
		}
		if cfg.Workspace.Writable == nil || *cfg.Workspace.Writable != true {
			t.Error("expected writable true by default")
		}
	})

	t.Run("no network defaults applied when empty", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{Command: []string{"test"}},
		}

		err := ApplyDefaults(cfg)
		if err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		// No default presets - air-gapped execution is valid
		if len(cfg.Network.Presets) != 0 {
			t.Errorf("expected no presets by default, got %v", cfg.Network.Presets)
		}
		if len(cfg.Network.Allow) != 0 {
			t.Errorf("expected no allow list by default, got %v", cfg.Network.Allow)
		}
	})

	t.Run("applies container security defaults", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{Command: []string{"test"}},
		}

		err := ApplyDefaults(cfg)
		if err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		if cfg.Container.NoNewPrivileges == nil || !*cfg.Container.NoNewPrivileges {
			t.Error("expected no_new_privileges to be true by default")
		}
	})
}

func TestExpandPathFromConfigTest(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		path string
		want string
	}{
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", home},
		{"~/subdir", filepath.Join(home, "subdir")},
		{"~/.config/file", filepath.Join(home, ".config/file")},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ExpandPath(tt.path)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// FindAgentfile tests are in discovery_test.go
