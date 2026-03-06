//go:build integration

package container

import (
	"context"
	"io"
	"testing"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/gbrindisi/agentbox/internal/config"
	"github.com/gbrindisi/agentbox/internal/output"
)

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

func TestContainerLifecycle(t *testing.T) {
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

	// Create test workspace
	workspace := t.TempDir()

	// Create config with simple echo command
	// Setting Network.Allow ensures ALLOWED_DOMAINS is set, which triggers
	// the firewall initialization code path.
	// NoNewPrivileges defaults to true - the new privilege model (root start + setpriv drop)
	// supports no_new_privileges: true, which is the more secure configuration.
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Command: []string{"echo"},
			Args:    []string{"hello"},
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

	// Create container
	containerID, err := CreateContainer(ctx, cli, cfg, nil, false, ImageTag())
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	defer Cleanup(ctx, cli, containerID)

	// Start container
	if err := cli.ContainerStart(ctx, containerID, dockercontainer.StartOptions{}); err != nil {
		t.Fatalf("ContainerStart() error = %v", err)
	}

	// Wait for container to exit
	exitCode, err := WaitContainer(ctx, cli, containerID)
	if err != nil {
		t.Fatalf("WaitContainer() error = %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Exit code = %d, want 0", exitCode)
	}
}

// TestNoNewPrivilegesTrue verifies that the container functions correctly with
// no_new_privileges: true (the default). This tests the new privilege model where:
// 1. Container starts as root
// 2. Firewall initialization runs as root (no sudo needed)
// 3. setpriv drops privileges to agent user
// 4. Agent command executes successfully
//
// This is a critical security test - no_new_privileges: true prevents privilege
// escalation via setuid binaries, and the new privilege model makes this work
// by design (privileges only flow downward, never upward).
func TestNoNewPrivilegesTrue(t *testing.T) {
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

	t.Run("container functions with no_new_privileges enabled", func(t *testing.T) {
		// Explicitly set no_new_privileges: true to verify the new privilege model works
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"echo"},
				Args:    []string{"no_new_privileges test passed"},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"example.com"}, // Triggers firewall initialization
			},
			Container: config.ContainerConfig{
				NoNewPrivileges: boolPtr(true), // Explicitly enabled
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

		// Wait for container to exit
		exitCode, err := WaitContainer(ctx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// The container should exit successfully:
		// - Firewall init completed (ran as root, no sudo needed)
		// - setpriv dropped privileges to agent
		// - Agent command executed successfully
		if exitCode != 0 {
			t.Errorf("Container should work with no_new_privileges: true, but got exit code %d", exitCode)
		}
	})

	t.Run("firewall initializes with no_new_privileges enabled", func(t *testing.T) {
		// This test verifies that firewall initialization works with no_new_privileges: true
		// by attempting to reach an allowed domain (which requires firewall setup)
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"curl"},
				Args:    []string{"-s", "--max-time", "10", "-o", "/dev/null", "https://example.com"},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"example.com"},
			},
			Container: config.ContainerConfig{
				NoNewPrivileges: boolPtr(true), // Explicitly enabled
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

		exitCode, err := WaitContainer(ctx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// curl should succeed - firewall allowed example.com
		if exitCode != 0 {
			t.Errorf("Firewall should work with no_new_privileges: true, but curl failed with exit code %d", exitCode)
		}
	})
}
