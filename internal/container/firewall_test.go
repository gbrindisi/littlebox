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

func TestFirewallBlocking(t *testing.T) {
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

	t.Run("blocked domain is unreachable", func(t *testing.T) {
		// Allow only api.anthropic.com, try to reach example.com (should fail)
		cfg := &config.Config{
			Agent: config.AgentConfig{
				// curl returns exit code 0 on success, non-zero on failure
				// --max-time 5 prevents hanging, -s for silent mode
				Command: []string{"curl"},
				Args:    []string{"-s", "--max-time", "5", "-o", "/dev/null", "https://example.com"},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				Allow: []string{"api.anthropic.com"},
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

		// curl should fail (non-zero exit) because example.com is blocked
		if exitCode == 0 {
			t.Error("Firewall should have blocked example.com, but curl succeeded")
		}
	})

	t.Run("allowed domain is reachable", func(t *testing.T) {
		// Allow example.com and try to reach it (should succeed)
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

		// curl should succeed (exit 0) because example.com is allowed
		if exitCode != 0 {
			t.Errorf("Allowed domain example.com should be reachable, but curl failed with exit code %d", exitCode)
		}
	})

	t.Run("DNS resolution works", func(t *testing.T) {
		// Test that DNS resolution works by resolving an allowed domain
		cfg := &config.Config{
			Agent: config.AgentConfig{
				// dig returns 0 on success, non-zero on failure
				Command: []string{"dig"},
				Args:    []string{"+short", "example.com"},
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

		exitCode, err := WaitContainer(ctx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// dig should succeed (exit 0) - DNS must work for the firewall to function
		if exitCode != 0 {
			t.Errorf("DNS resolution should work, but dig failed with exit code %d", exitCode)
		}
	})
}
