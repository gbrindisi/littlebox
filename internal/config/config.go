// Package config provides configuration parsing and management for agentbox.
package config

// Config represents the complete configuration for agentbox.
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
