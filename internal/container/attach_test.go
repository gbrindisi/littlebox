//go:build integration

package container

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/client"

	"github.com/gbrindisi/littlebox/internal/config"
	"github.com/gbrindisi/littlebox/internal/output"
)

// TestAttachContainerExitsCleanly verifies that AttachContainer returns promptly
// when the container exits, without blocking on stdin. This was a bug where
// the code would wait for stdin to close even after output completed.
func TestAttachContainerExitsCleanly(t *testing.T) {
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
	if err := mgr.EnsureImage(ctx, false, output.Debug, io.Discard); err != nil {
		t.Fatalf("EnsureImage() error = %v", err)
	}

	workspace := t.TempDir()

	t.Run("returns promptly when container exits without reading stdin", func(t *testing.T) {
		// Create a container that exits immediately without reading stdin
		// This simulates `claude --print` which outputs and exits
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"echo"},
				Args:    []string{"hello world"},
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

		// Use a timeout to detect if AttachContainer hangs
		done := make(chan error, 1)
		go func() {
			done <- AttachContainer(ctx, cli, containerID, false, output.Debug, nil)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("AttachContainer() error = %v", err)
			}
			// Success - returned promptly
		case <-time.After(10 * time.Second):
			t.Fatal("AttachContainer() timed out - likely blocked waiting for stdin")
		}
	})

	t.Run("returns promptly with TTY mode", func(t *testing.T) {
		// Same test but with TTY enabled
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"echo"},
				Args:    []string{"hello with tty"},
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

		containerID, err := CreateContainer(ctx, cli, cfg, nil, true, ImageTag())
		if err != nil {
			t.Fatalf("CreateContainer() error = %v", err)
		}
		defer Cleanup(ctx, cli, containerID)

		done := make(chan error, 1)
		go func() {
			done <- AttachContainer(ctx, cli, containerID, true, output.Debug, nil)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("AttachContainer() error = %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("AttachContainer() with TTY timed out - likely blocked waiting for stdin")
		}
	})

	t.Run("handles command that reads stdin then exits", func(t *testing.T) {
		// Container that reads a bit of stdin then exits
		// This tests that the stdin goroutine doesn't cause issues
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"head"},
				Args:    []string{"-c", "0"}, // Read 0 bytes, exit immediately
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

		done := make(chan error, 1)
		go func() {
			done <- AttachContainer(ctx, cli, containerID, false, output.Debug, nil)
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("AttachContainer() error = %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("AttachContainer() timed out with stdin-reading command")
		}
	})
}
