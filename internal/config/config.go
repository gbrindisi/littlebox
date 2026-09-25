// Package config provides configuration parsing and management for littlebox.
package config

// Config represents the complete configuration for littlebox.
type Config struct {
	Agent     AgentConfig     `yaml:"agent"`
	Workspace WorkspaceConfig `yaml:"workspace"`
	Mounts    []MountConfig   `yaml:"mounts,omitempty"`
	Redact    []string        `yaml:"redact,omitempty"`
	Network   NetworkConfig   `yaml:"network"`
	Container ContainerConfig `yaml:"container"`
	Env       []EnvVar        `yaml:"env,omitempty"`
}

// AgentConfig defines the agent to run inside the container.
type AgentConfig struct {
	Command        []string `yaml:"command,omitempty"`
	Args           []string `yaml:"args,omitempty"`
	Env            []EnvVar `yaml:"env,omitempty"`
	EnvPassthrough []string `yaml:"env_passthrough,omitempty"`
	BuildScript    string   `yaml:"build_script,omitempty"`
	// ImageScope controls derived image caching: "workspace" (default) builds a
	// separate image per workspace path; "shared" reuses one image per build script.
	ImageScope string `yaml:"image_scope,omitempty"`
}

// Image scope values for AgentConfig.ImageScope.
const (
	ImageScopeWorkspace = "workspace"
	ImageScopeShared    = "shared"
)

// ImageScopeKey returns the workspace component used when hashing the derived
// image tag. It is empty for the shared scope so all workspaces share one image.
func (c *Config) ImageScopeKey() string {
	if c.Agent.ImageScope == ImageScopeShared {
		return ""
	}
	return c.Workspace.Path
}

// WorkspaceConfig defines how the workspace is mounted.
type WorkspaceConfig struct {
	Path       string `yaml:"path,omitempty"`
	MountPoint string `yaml:"mount_point,omitempty"`
	Writable   *bool  `yaml:"writable,omitempty"`
}

// MountConfig defines an additional mount from host to container.
type MountConfig struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"readonly,omitempty"`
}

// NetworkConfig defines network access rules.
type NetworkConfig struct {
	Presets     []string `yaml:"presets,omitempty"`
	Allow       []string `yaml:"allow,omitempty"`
	Deny        []string `yaml:"deny,omitempty"`
	DenyAllElse bool     `yaml:"deny_all_else,omitempty"`

	// Preset is deprecated - used only to detect old config format and provide migration message.
	// Use Presets (plural) instead.
	Preset string `yaml:"preset,omitempty"`
}

// ContainerConfig defines container lifecycle behavior.
type ContainerConfig struct {
	Keep            bool   `yaml:"keep,omitempty"`
	ReadOnlyRoot    bool   `yaml:"readonly_root,omitempty"`
	NoNewPrivileges *bool  `yaml:"no_new_privileges,omitempty"`
	SeccompProfile  string `yaml:"seccomp_profile,omitempty"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value,omitempty"`
}
