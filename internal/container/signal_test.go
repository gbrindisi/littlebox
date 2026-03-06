//go:build integration

package container

import (
	"context"
	"io"
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/gbrindisi/agentbox/internal/config"
	"github.com/gbrindisi/agentbox/internal/output"
)

func TestSignalForwarding(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	// Verify Docker daemon is reachable
	_, err = cli.Ping(ctx)
	if err != nil {
		t.Skip("Docker daemon not reachable:", err)
	}

	// Create manager for image operations
	mgr := NewManagerWithClient(cli)

	// Build or use cached image
	if err := mgr.EnsureImage(ctx, false, output.Quiet, io.Discard); err != nil {
		t.Fatalf("EnsureImage() error = %v", err)
	}

	workspace := t.TempDir()

	t.Run("container responds to SIGTERM", func(t *testing.T) {
		// Use a bash script that handles SIGTERM properly
		// The sleep command alone ignores SIGTERM, so we need trap
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"bash"},
				Args: []string{"-c", `
					trap 'echo "Received SIGTERM"; exit 0' TERM
					while true; do sleep 1; done
				`},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"example.com"}, // Need at least one domain for firewall init
			},
		}
		if err := config.ApplyDefaults(cfg); err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		containerID, err := CreateContainer(ctx, cli, cfg, nil, false, ImageTag())
		if err != nil {
			t.Fatalf("CreateContainer() error = %v", err)
		}
		defer Cleanup(ctx, cli, containerID)

		// Start container
		if err := cli.ContainerStart(ctx, containerID, dockercontainer.StartOptions{}); err != nil {
			t.Fatalf("ContainerStart() error = %v", err)
		}

		// Give the container time to start
		time.Sleep(3 * time.Second)

		// Record time before sending signal
		startTime := time.Now()

		// Send SIGTERM to the container
		if err := cli.ContainerKill(ctx, containerID, "SIGTERM"); err != nil {
			t.Fatalf("ContainerKill(SIGTERM) error = %v", err)
		}

		// Wait for container to exit with timeout
		waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		exitCode, err := WaitContainer(waitCtx, cli, containerID)
		elapsed := time.Since(startTime)

		if err != nil {
			t.Fatalf("WaitContainer() error = %v (container did not exit within timeout)", err)
		}

		// The container should exit quickly (within a few seconds) after SIGTERM
		if elapsed > 5*time.Second {
			t.Errorf("Container took too long to exit after SIGTERM: %v", elapsed)
		}

		// Exit code should be 0 (we explicitly exit 0 in trap)
		if exitCode != 0 {
			t.Logf("Exit code = %d (expected 0 from trap handler)", exitCode)
		}

		t.Logf("Container exited with code %d after SIGTERM in %v", exitCode, elapsed)
	})

	t.Run("exit code reflects signal", func(t *testing.T) {
		// Use a shell script that explicitly exits with signal-based code
		cfg := &config.Config{
			Agent: config.AgentConfig{
				// Use bash to trap SIGTERM and exit with code 143 (128 + 15)
				Command: []string{"bash"},
				Args: []string{"-c", `
					trap 'exit 143' TERM
					while true; do sleep 1; done
				`},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"example.com"},
			},
		}
		if err := config.ApplyDefaults(cfg); err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		containerID, err := CreateContainer(ctx, cli, cfg, nil, false, ImageTag())
		if err != nil {
			t.Fatalf("CreateContainer() error = %v", err)
		}
		defer Cleanup(ctx, cli, containerID)

		if err := cli.ContainerStart(ctx, containerID, dockercontainer.StartOptions{}); err != nil {
			t.Fatalf("ContainerStart() error = %v", err)
		}

		time.Sleep(3 * time.Second)

		// Send SIGTERM
		if err := cli.ContainerKill(ctx, containerID, "SIGTERM"); err != nil {
			t.Fatalf("ContainerKill(SIGTERM) error = %v", err)
		}

		waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		exitCode, err := WaitContainer(waitCtx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// With explicit trap, exit code should be 143
		if exitCode != 143 {
			t.Errorf("Exit code = %d, want 143 (128 + SIGTERM)", exitCode)
		}
	})

	t.Run("cleanup completes after signal", func(t *testing.T) {
		// This test verifies that container cleanup works properly after signal termination
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"bash"},
				Args: []string{"-c", `
					trap 'exit 0' TERM
					while true; do sleep 1; done
				`},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"example.com"},
			},
		}
		if err := config.ApplyDefaults(cfg); err != nil {
			t.Fatalf("ApplyDefaults() error = %v", err)
		}

		containerID, err := CreateContainer(ctx, cli, cfg, nil, false, ImageTag())
		if err != nil {
			t.Fatalf("CreateContainer() error = %v", err)
		}

		if err := cli.ContainerStart(ctx, containerID, dockercontainer.StartOptions{}); err != nil {
			t.Fatalf("ContainerStart() error = %v", err)
		}

		time.Sleep(3 * time.Second)

		// Send SIGTERM
		if err := cli.ContainerKill(ctx, containerID, "SIGTERM"); err != nil {
			t.Fatalf("ContainerKill(SIGTERM) error = %v", err)
		}

		// Wait for exit with timeout
		waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		_, err = WaitContainer(waitCtx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// Cleanup should complete without error (idempotent - may be auto-removed already)
		Cleanup(ctx, cli, containerID)

		// Give Docker a moment to auto-remove the container
		time.Sleep(500 * time.Millisecond)

		// Verify container is gone by trying to inspect it
		_, err = cli.ContainerInspect(ctx, containerID)
		if err == nil {
			t.Error("Container should have been removed after cleanup")
		} else {
			// Expected - container was removed (either by AutoRemove or Cleanup)
			t.Log("Container cleanup completed successfully")
		}
	})
}
