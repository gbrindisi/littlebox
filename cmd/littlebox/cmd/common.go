package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gbrindisi/littlebox/internal/config"
	"github.com/gbrindisi/littlebox/internal/container"
	"github.com/gbrindisi/littlebox/internal/output"
)

// runOptions configures how runContainer executes.
type runOptions struct {
	cfg     *config.Config
	args    []string
	ttyMode container.TTYMode
	rebuild bool
	debug   bool
}

// runContainer handles the full container lifecycle: create manager, ensure image,
// create container, setup signals, handle resize, attach, and wait.
// Returns the container's exit code and any error that occurred.
func runContainer(ctx context.Context, opts runOptions) (int, error) {
	mgr, err := container.NewManager()
	if err != nil {
		return 1, fmt.Errorf("failed to create container manager: %w", err)
	}
	defer func() { _ = mgr.Close() }()

	// Determine verbosity from debug flag
	verbosity := output.Quiet
	if opts.debug {
		verbosity = output.Debug
	}

	if err := mgr.EnsureImage(ctx, false, verbosity, os.Stdout); err != nil {
		return 1, fmt.Errorf("failed to ensure image: %w", err)
	}

	// Determine which image to use: derived image if BuildScript is set, otherwise base image
	imageTag := container.ImageTag()
	if opts.cfg.Agent.BuildScript != "" {
		derivedTag, err := mgr.EnsureDerivedImage(ctx, opts.cfg.Agent.BuildScript, opts.cfg.Workspace.Path, opts.rebuild, verbosity, os.Stderr)
		if err != nil {
			return 1, fmt.Errorf("failed to ensure derived image: %w", err)
		}
		imageTag = derivedTag
	}

	tty := container.DetectTTY(opts.ttyMode)

	containerID, err := container.CreateContainer(ctx, mgr.Client(), opts.cfg, opts.args, tty, imageTag)
	if err != nil {
		return 1, fmt.Errorf("failed to create container: %w", err)
	}
	defer container.Cleanup(ctx, mgr.Client(), containerID)

	cleanupSignals := container.ForwardSignals(ctx, mgr.Client(), containerID)
	defer cleanupSignals()

	if tty {
		cleanupResize := container.HandleResize(ctx, mgr.Client(), containerID)
		defer cleanupResize()
	}

	// Print sandbox setup status in quiet mode
	statusWriter := output.NewStatusWriter(os.Stdout, verbosity)
	statusWriter.Start("Setting up the sandbox")

	// statusDone callback prints "done" when setup completes
	statusDone := func() {
		statusWriter.Done()
		output.BulletPrint(os.Stdout, verbosity, "Running the agent")
	}

	if err := container.AttachContainer(ctx, mgr.Client(), containerID, tty, verbosity, statusDone); err != nil {
		return 1, fmt.Errorf("failed to attach to container: %w", err)
	}

	exitCode, _ := container.WaitContainer(ctx, mgr.Client(), containerID)
	return exitCode, nil
}

// loadAndValidateConfig loads configuration from the given path (or auto-discovers),
// applies workspace override, defaults, and validates.
func loadAndValidateConfig() (*config.Config, error) {
	cfgPath := configFile
	if cfgPath == "" {
		// Try to find Agentfile in current or parent directories
		var err error
		cfgPath, err = config.FindAgentfile("")
		if err != nil {
			// If not found, use default which will error with a better message
			cfgPath = config.DefaultConfigFile
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override workspace if flag provided
	if workspace != "" {
		cfg.Workspace.Path = workspace
	}

	// Apply defaults and expand paths relative to Agentfile directory
	if err := config.ApplyDefaults(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply defaults: %w", err)
	}
	cfg.ExpandPaths(filepath.Dir(cfgPath))

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	return cfg, nil
}
