package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

// ValidationWarning represents a non-fatal configuration warning.
type ValidationWarning struct {
	Message string
}

// ValidationWarnings is a collection of validation warnings.
type ValidationWarnings []ValidationWarning

// Validate checks the configuration for errors and warnings.
// It assumes ApplyDefaults has already been called.
// Returns validation errors (if any) and warnings (which don't prevent execution).
func Validate(cfg *Config) error {
	var errs ValidationErrors
	var warns ValidationWarnings

	errs = append(errs, validateAgent(&cfg.Agent)...)
	errs = append(errs, validateWorkspace(&cfg.Workspace)...)
	errs = append(errs, validateMounts(cfg.Mounts)...)

	networkErrs, networkWarns := validateNetwork(&cfg.Network)
	errs = append(errs, networkErrs...)
	warns = append(warns, networkWarns...)

	errs = append(errs, validateEnvPassthrough(cfg.Agent.EnvPassthrough)...)

	// Print warnings to stderr (they don't prevent execution)
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "[agentbox] warning: %s\n", w.Message)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateAgent(agent *AgentConfig) ValidationErrors {
	var errs ValidationErrors

	// Command is required since we no longer have profiles
	if len(agent.Command) == 0 {
		errs = append(errs, ValidationError{
			Field:   "agent.command",
			Message: "agent.command is required",
		})
	}

	return errs
}

func validateWorkspace(ws *WorkspaceConfig) ValidationErrors {
	var errs ValidationErrors

	if ws.Path != "" {
		info, err := os.Stat(ws.Path)
		if os.IsNotExist(err) {
			errs = append(errs, ValidationError{
				Field:   "workspace.path",
				Message: fmt.Sprintf("directory does not exist: %s", ws.Path),
			})
		} else if err != nil {
			errs = append(errs, ValidationError{
				Field:   "workspace.path",
				Message: fmt.Sprintf("cannot access directory: %v", err),
			})
		} else if !info.IsDir() {
			errs = append(errs, ValidationError{
				Field:   "workspace.path",
				Message: fmt.Sprintf("not a directory: %s", ws.Path),
			})
		}
	}

	return errs
}

func validateMounts(mounts []MountConfig) ValidationErrors {
	var errs ValidationErrors

	for i, m := range mounts {
		if m.Source == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mounts[%d].source", i),
				Message: "source path is required",
			})
			continue
		}

		if m.Target == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mounts[%d].target", i),
				Message: "target path is required",
			})
		}

		if _, err := os.Stat(m.Source); os.IsNotExist(err) {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("mounts[%d].source", i),
				Message: fmt.Sprintf("path does not exist: %s", m.Source),
			})
		}
	}

	return errs
}

func validateNetwork(net *NetworkConfig) (ValidationErrors, ValidationWarnings) {
	var errs ValidationErrors
	var warns ValidationWarnings

	// Check for old preset: (singular) format - migration error
	if net.Preset != "" {
		errs = append(errs, ValidationError{
			Field: "network.preset",
			Message: fmt.Sprintf("'preset' (singular) is no longer supported. Migrate to 'presets' (plural) list format.\n"+
				"  Old format:  preset: %s\n"+
				"  New format:  presets: [anthropic, github, npm]  # choose the services you need\n"+
				"  Available presets: %v", net.Preset, sortedPresetNames()),
		})
	}

	// Validate each preset name in the list
	for _, preset := range net.Presets {
		if _, ok := NetworkPresets[preset]; !ok {
			errs = append(errs, ValidationError{
				Field:   "network.presets",
				Message: fmt.Sprintf("unknown preset %q (available: %v)", preset, sortedPresetNames()),
			})
		}
	}

	// Warn when neither presets nor allow is configured (air-gapped is valid but user should be aware)
	if len(net.Presets) == 0 && len(net.Allow) == 0 && net.Preset == "" {
		warns = append(warns, ValidationWarning{
			Message: "No network presets or allow list configured. Container will have no outbound network access (except DNS).",
		})
	}

	return errs, warns
}

// sortedPresetNames returns all valid preset names in sorted order.
func sortedPresetNames() []string {
	names := make([]string, 0, len(NetworkPresets))
	for k := range NetworkPresets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func validateEnvPassthrough(vars []string) ValidationErrors {
	var errs ValidationErrors

	for _, v := range vars {
		if isGlobPattern(v) {
			continue // Skip glob patterns - they match zero or more variables
		}
		if os.Getenv(v) == "" {
			errs = append(errs, ValidationError{
				Field:   "agent.env_passthrough",
				Message: fmt.Sprintf("required environment variable %s is not set", v),
			})
		}
	}

	return errs
}
