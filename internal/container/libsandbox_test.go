//go:build integration

package container

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/gbrindisi/littlebox/internal/config"
)

// getContainerLogs retrieves both stdout and stderr from a container
func getContainerLogs(ctx context.Context, cli *client.Client, containerID string) (string, error) {
	logs, err := cli.ContainerLogs(ctx, containerID, dockercontainer.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: false,
	})
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf, err := io.ReadAll(logs)
	if err != nil {
		return "", err
	}

	// Docker logs are in a special format with 8-byte headers
	// We need to strip these headers to get the actual log content
	var output strings.Builder
	for len(buf) > 0 {
		if len(buf) < 8 {
			break
		}
		// Header format: [stream type][0][0][0][size as 4 bytes]
		// Skip the header and read the payload
		size := int(buf[4])<<24 | int(buf[5])<<16 | int(buf[6])<<8 | int(buf[7])
		if len(buf) < 8+size {
			break
		}
		output.Write(buf[8 : 8+size])
		buf = buf[8+size:]
	}

	return output.String(), nil
}

func TestLibsandboxBlockedConnection(t *testing.T) {
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

	// Force rebuild the image with no cache to ensure we have the latest libsandbox.so and init-firewall.sh
	if err := mgr.BuildImage(ctx, &BuildOptions{
		ImageName: ImageTag(),
		NoCache:   true,
		Output:    io.Discard,
	}); err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	workspace := t.TempDir()

	t.Run("blocked connection shows custom error message", func(t *testing.T) {
		// Allow only api.anthropic.com, try to reach example.com (should fail)
		// First check if LD_PRELOAD is set and test libsandbox
		cfg := &config.Config{
			Agent: config.AgentConfig{
				Command: []string{"sh"},
				Args: []string{"-c", `
echo "LD_PRELOAD=$LD_PRELOAD"
echo "Checking if libsandbox.so exists:"
ls -la /usr/local/lib/libsandbox.so
echo "Checking if curl is dynamically linked:"
ldd /usr/bin/curl | grep -i libc
echo "Now testing connection to example.com:"
curl --max-time 5 https://example.com 2>&1
`},
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

		// Wait for container to finish
		exitCode, err := WaitContainer(ctx, cli, containerID)
		if err != nil {
			t.Fatalf("WaitContainer() error = %v", err)
		}

		// curl should fail (non-zero exit) because example.com is blocked
		if exitCode == 0 {
			t.Error("Expected curl to fail when connecting to blocked domain")
		}

		// Get container logs to check stderr
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logs, err := getContainerLogs(ctx2, cli, containerID)
		if err != nil {
			t.Fatalf("Failed to get container logs: %v", err)
		}

		// Verify the custom error message appears in logs
		if !strings.Contains(logs, "agentbox: Connection to") {
			t.Errorf("Expected custom error message starting with 'agentbox: Connection to', got logs:\n%s", logs)
		}
		if !strings.Contains(logs, "blocked by sandbox firewall") {
			t.Errorf("Expected error message to mention 'blocked by sandbox firewall', got logs:\n%s", logs)
		}
		if !strings.Contains(logs, "This is not bypassable") {
			t.Errorf("Expected error message to state 'This is not bypassable', got logs:\n%s", logs)
		}
	})

	t.Run("allowed connection passes without error message", func(t *testing.T) {
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
			t.Errorf("Expected curl to succeed when connecting to allowed domain, got exit code %d", exitCode)
		}

		// Get container logs to verify no sandbox error message
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logs, err := getContainerLogs(ctx2, cli, containerID)
		if err != nil {
			t.Fatalf("Failed to get container logs: %v", err)
		}

		// Verify NO sandbox error message appears (should pass through silently)
		if strings.Contains(logs, "agentbox: Connection to") {
			t.Errorf("Expected no sandbox error message for allowed connection, but got:\n%s", logs)
		}
	})

	t.Run("CIDR range matching works correctly", func(t *testing.T) {
		// Test that IPs within a CIDR range are allowed
		// Allow a wide CIDR range and verify connections to IPs within it work
		cfg := &config.Config{
			Agent: config.AgentConfig{
				// Use curl to connect to specific IPs that should be in the CIDR range
				// Cloudflare uses 104.18.0.0/15 range, so 104.18.26.120 should match 104.18.0.0/16
				Command: []string{"sh"},
				Args: []string{"-c", `
					# Test connecting to an IP that's in the allowed CIDR range
					# example.com often resolves to 104.18.x.x which is in Cloudflare's range
					IP="104.18.26.120"
					echo "Testing connection to $IP (should be in allowed CIDR 104.18.0.0/16)..."
					curl --max-time 10 -s -o /dev/null http://$IP && echo "SUCCESS" || echo "FAILED"
				`},
			},
			Workspace: config.WorkspaceConfig{
				Path: workspace,
			},
			Network: config.NetworkConfig{
				// Allow a /16 CIDR block that covers Cloudflare IPs
				Allow: []string{"104.18.0.0/16"},
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

		// Get logs to see what happened
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logs, err := getContainerLogs(ctx2, cli, containerID)
		if err != nil {
			t.Fatalf("Failed to get container logs: %v", err)
		}

		// The script should succeed if CIDR matching works
		if !strings.Contains(logs, "SUCCESS") {
			t.Errorf("Expected CIDR range matching to allow connection, got logs:\n%s\nExit code: %d", logs, exitCode)
		}

		// Verify no sandbox error message for allowed CIDR
		if strings.Contains(logs, "agentbox: Connection to") && strings.Contains(logs, "blocked") {
			t.Errorf("Expected no block message for IP in allowed CIDR range, got:\n%s", logs)
		}
	})
}
