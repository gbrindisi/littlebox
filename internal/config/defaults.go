package config

const (
	// DefaultWorkspaceMountPoint is where the workspace is mounted in the container.
	DefaultWorkspaceMountPoint = "/workspace"
)

// ApplyDefaults fills in default values for unset configuration fields.
// It modifies the config in place.
func ApplyDefaults(cfg *Config) error {
	if err := applyAgentDefaults(cfg); err != nil {
		return err
	}

	if err := applyWorkspaceDefaults(cfg); err != nil {
		return err
	}

	if err := applyMountDefaults(cfg); err != nil {
		return err
	}

	if err := applyNetworkDefaults(cfg); err != nil {
		return err
	}

	applyContainerDefaults(cfg)
	return nil
}

func applyAgentDefaults(cfg *Config) error {
	if cfg.Agent.ImageScope == "" {
		cfg.Agent.ImageScope = ImageScopeWorkspace
	}
	return nil
}

func applyWorkspaceDefaults(cfg *Config) error {
	if cfg.Workspace.Path == "" {
		// Default to "." which will be resolved relative to Agentfile directory
		// by ExpandPaths(). This makes configurations portable - an Agentfile
		// can be moved with its project and still work without modification.
		cfg.Workspace.Path = "."
	}
	// Note: Workspace path tilde expansion is handled by ExpandPaths()

	if cfg.Workspace.MountPoint == "" {
		cfg.Workspace.MountPoint = DefaultWorkspaceMountPoint
	}

	if cfg.Workspace.Writable == nil {
		writable := true
		cfg.Workspace.Writable = &writable
	}

	return nil
}

func applyMountDefaults(cfg *Config) error {
	// Mount path tilde expansion is handled by ExpandPaths()
	return nil
}

func applyNetworkDefaults(cfg *Config) error {
	// No default presets - if neither presets nor allow is configured,
	// the container will have no outbound network access (except DNS).
	// This is a valid use case for air-gapped execution.
	return ApplyNetworkPresets(cfg)
}

func applyContainerDefaults(cfg *Config) {
	// Enable no-new-privileges by default for security hardening
	// This prevents privilege escalation via setuid/setgid binaries
	// Only set default if not explicitly configured
	if cfg.Container.NoNewPrivileges == nil {
		noNewPrivileges := true
		cfg.Container.NoNewPrivileges = &noNewPrivileges
	}
}

// ExpandPath expands ~ to the user's home directory for use in other packages.
// Only ~ at the start of a path followed by / or end of string is expanded.
// ~user/path is NOT expanded (only ~ alone).
func ExpandPath(path string) string {
	return expandTilde(path)
}
