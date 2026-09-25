package config

import (
	"os"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("valid config passes", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
			Network: NetworkConfig{
				Presets: []string{"anthropic", "github"},
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("missing command fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error when command not specified")
		}
	})

	t.Run("nonexistent workspace path fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: "/nonexistent/path/that/does/not/exist",
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for nonexistent workspace path")
		}
	})

	t.Run("workspace path not a directory fails", func(t *testing.T) {
		// Create a file, not a directory
		tmpFile, err := os.CreateTemp(tmpDir, "not-a-dir")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		_ = tmpFile.Close()

		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: tmpFile.Name(),
			},
		}

		err = Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error when workspace path is not a directory")
		}
	})

	t.Run("mount missing source fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Mounts: []MountConfig{
				{Target: "/container/path"},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for mount without source")
		}
	})

	t.Run("mount missing target fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Mounts: []MountConfig{
				{Source: tmpDir},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for mount without target")
		}
	})

	t.Run("mount nonexistent source fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Mounts: []MountConfig{
				{Source: "/nonexistent/path", Target: "/target"},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for mount with nonexistent source")
		}
	})

	t.Run("unknown network preset fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Network: NetworkConfig{
				Presets: []string{"unknown-preset"},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for unknown network preset")
		}
		// Verify error message lists available presets
		errMsg := err.Error()
		if !strings.Contains(errMsg, "unknown preset") {
			t.Errorf("error should mention 'unknown preset', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "anthropic") {
			t.Errorf("error should list available presets like 'anthropic', got: %s", errMsg)
		}
	})

	t.Run("old preset format fails with migration message", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Network: NetworkConfig{
				Preset: "standard", // Old singular format
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for old preset format")
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "preset") || !strings.Contains(errMsg, "singular") {
			t.Errorf("error should mention old singular format, got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "presets") {
			t.Errorf("error should mention new plural format, got: %s", errMsg)
		}
	})

	t.Run("valid presets pass", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
			Network: NetworkConfig{
				Presets: []string{"anthropic", "github", "npm"},
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("Validate() unexpected error for valid presets: %v", err)
		}
	})

	t.Run("empty network config warns but passes", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
			Network: NetworkConfig{
				// No presets, no allow - air-gapped mode
			},
		}

		// Capture stderr for warning check
		err := Validate(cfg)
		// Should NOT return an error - warnings don't fail validation
		if err != nil {
			t.Errorf("Validate() should pass for empty network config (air-gapped), got error: %v", err)
		}
		// Note: warning is printed to stderr, not returned as error
	})

	t.Run("only allow list passes without warning", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command: []string{"test"},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
			Network: NetworkConfig{
				Allow: []string{"custom.example.com"},
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("Validate() unexpected error for allow-only config: %v", err)
		}
	})

	t.Run("missing optional env passthrough var passes", func(t *testing.T) {
		envVar := "TEST_NONEXISTENT_VAR_12345"
		_ = os.Unsetenv(envVar)

		cfg := &Config{
			Agent: AgentConfig{
				Command:        []string{"test"},
				EnvPassthrough: []EnvPassthroughEntry{{Name: envVar}},
			},
			Workspace: WorkspaceConfig{Path: tmpDir},
		}

		if err := Validate(cfg); err != nil {
			t.Errorf("Validate() unexpected error for optional missing var: %v", err)
		}
	})

	t.Run("glob pattern marked required fails", func(t *testing.T) {
		cfg := &Config{
			Agent: AgentConfig{
				Command:        []string{"test"},
				EnvPassthrough: []EnvPassthroughEntry{{Name: "CLAUDE_CODE_*", Required: true}},
			},
			Workspace: WorkspaceConfig{Path: tmpDir},
		}

		err := Validate(cfg)
		if err == nil || !strings.Contains(err.Error(), "cannot be marked required") {
			t.Errorf("Validate() expected glob+required error, got %v", err)
		}
	})

	t.Run("missing required env passthrough var fails", func(t *testing.T) {
		// Ensure the env var is not set
		envVar := "TEST_NONEXISTENT_VAR_12345"
		_ = os.Unsetenv(envVar)

		cfg := &Config{
			Agent: AgentConfig{
				Command:        []string{"test"},
				EnvPassthrough: []EnvPassthroughEntry{{Name: envVar, Required: true}},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("Validate() expected error for missing env passthrough variable")
		}
	})

	t.Run("set env passthrough var passes", func(t *testing.T) {
		envVar := "TEST_VAR_SET_12345"
		_ = os.Setenv(envVar, "test-value")
		defer func() { _ = os.Unsetenv(envVar) }()

		cfg := &Config{
			Agent: AgentConfig{
				Command:        []string{"test"},
				EnvPassthrough: []EnvPassthroughEntry{{Name: envVar, Required: true}},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})

	t.Run("glob pattern env passthrough skips validation", func(t *testing.T) {
		// Glob patterns like CLAUDE_CODE_* should not be validated via os.Getenv
		// since they match zero or more variables at runtime
		cfg := &Config{
			Agent: AgentConfig{
				Command:        []string{"test"},
				EnvPassthrough: []EnvPassthroughEntry{{Name: "CLAUDE_CODE_*"}, {Name: "TEST_PREFIX_???"}, {Name: "VAR_[A-Z]"}},
			},
			Workspace: WorkspaceConfig{
				Path: tmpDir,
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("Validate() unexpected error for glob patterns: %v", err)
		}
	})

	t.Run("collects all errors together", func(t *testing.T) {
		// Create a config with multiple validation errors
		cfg := &Config{
			Agent: AgentConfig{
				// no command - error 1
			},
			Workspace: WorkspaceConfig{
				Path: "/nonexistent/path/12345", // error 2: nonexistent path
			},
			Network: NetworkConfig{
				Presets: []string{"invalid-preset"}, // error 3: unknown preset
			},
			Mounts: []MountConfig{
				{Source: "/nonexistent/mount/path", Target: "/target"}, // error 4: nonexistent mount
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Fatal("Validate() expected errors")
		}

		// Should be ValidationErrors type with multiple errors
		valErrs, ok := err.(ValidationErrors)
		if !ok {
			t.Fatalf("expected ValidationErrors, got %T", err)
		}

		// Should have at least 4 errors (command, workspace, network, mount)
		if len(valErrs) < 4 {
			t.Errorf("expected at least 4 errors, got %d: %v", len(valErrs), valErrs)
		}
	})
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errs := ValidationErrors{}
		if errs.Error() != "" {
			t.Errorf("expected empty string, got %q", errs.Error())
		}
	})

	t.Run("single error", func(t *testing.T) {
		errs := ValidationErrors{
			{Field: "test.field", Message: "test message"},
		}
		got := errs.Error()
		if got == "" {
			t.Error("expected non-empty error string")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := ValidationErrors{
			{Field: "field1", Message: "message1"},
			{Field: "field2", Message: "message2"},
		}
		got := errs.Error()
		if got == "" {
			t.Error("expected non-empty error string")
		}
	})
}
