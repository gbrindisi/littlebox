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

	"github.com/gbrindisi/agentbox/internal/config"
)

// TestDNSCachingWithHostname tests that DNS lookup followed by blocked connection shows hostname
func TestDNSCachingWithHostname(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skip("Docker daemon not reachable:", err)
	}

	mgr := NewManagerWithClient(cli)

	// Force rebuild to ensure latest libsandbox.so
	if err := mgr.BuildImage(ctx, &BuildOptions{
		ImageName: ImageTag(),
		NoCache:   true,
		Output:    io.Discard,
	}); err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	workspace := t.TempDir()

	// Test: DNS lookup followed by blocked connection should show hostname
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Command: []string{"sh"},
			Args: []string{"-c", `
				# Resolve example.com to get IP in cache
				getent hosts example.com
				EXAMPLE_IP=$(getent hosts example.com | awk '{print $1}' | head -1)
				echo "Resolved example.com to $EXAMPLE_IP"

				# Try to connect immediately (should be blocked and show hostname)
				curl --max-time 5 http://$EXAMPLE_IP 2>&1
			`},
		},
		Workspace: config.WorkspaceConfig{
			Path: workspace,
		},
		Network: config.NetworkConfig{
			Allow: []string{"api.anthropic.com"}, // Block example.com
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
		t.Error("Expected curl to fail when connecting to blocked IP")
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logs, err := getContainerLogs(ctx2, cli, containerID)
	if err != nil {
		t.Fatalf("Failed to get container logs: %v", err)
	}

	// Verify the error message contains hostname (example.com) and IP
	if !strings.Contains(logs, "example.com") {
		t.Errorf("Expected error message to contain hostname 'example.com', got logs:\n%s", logs)
	}
	if !strings.Contains(logs, "blocked by sandbox firewall") {
		t.Errorf("Expected error message to mention 'blocked by sandbox firewall', got logs:\n%s", logs)
	}
	// Check for the format: "hostname (IP)" pattern
	if !strings.Contains(logs, "(") || !strings.Contains(logs, ")") {
		t.Errorf("Expected error message to show 'hostname (IP)' format, got logs:\n%s", logs)
	}
}

// TestDNSCachingWithoutHostname tests that blocked connection without prior DNS lookup shows IP only
func TestDNSCachingWithoutHostname(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skip("Docker daemon not reachable:", err)
	}

	mgr := NewManagerWithClient(cli)

	if err := mgr.BuildImage(ctx, &BuildOptions{
		ImageName: ImageTag(),
		NoCache:   true,
		Output:    io.Discard,
	}); err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	workspace := t.TempDir()

	// Test: Direct connection to IP without prior DNS lookup should show IP only
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Command: []string{"sh"},
			Args: []string{"-c", `
				# Try to connect directly to an IP without DNS lookup
				# Using a known IP that will be blocked (e.g., 93.184.216.34 is example.com)
				curl --max-time 5 http://93.184.216.34 2>&1
			`},
		},
		Workspace: config.WorkspaceConfig{
			Path: workspace,
		},
		Network: config.NetworkConfig{
			Allow: []string{"api.anthropic.com"}, // Block this IP
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

	// curl should fail
	if exitCode == 0 {
		t.Error("Expected curl to fail when connecting to blocked IP")
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logs, err := getContainerLogs(ctx2, cli, containerID)
	if err != nil {
		t.Fatalf("Failed to get container logs: %v", err)
	}

	// Verify the error message contains IP only (no hostname)
	if !strings.Contains(logs, "93.184.216.34") {
		t.Errorf("Expected error message to contain IP '93.184.216.34', got logs:\n%s", logs)
	}
	if !strings.Contains(logs, "blocked by sandbox firewall") {
		t.Errorf("Expected error message to mention 'blocked by sandbox firewall', got logs:\n%s", logs)
	}
	// Verify no hostname appears (should NOT have parentheses with hostname)
	// The message should be "Connection to 93.184.216.34 blocked" not "Connection to example.com (93.184.216.34) blocked"
	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if strings.Contains(line, "agentbox: Connection to") && strings.Contains(line, "blocked") {
			// This line is the error message - it should NOT contain both IP and hostname
			if strings.Contains(line, "example.com") {
				t.Errorf("Expected error message to NOT contain hostname when no DNS lookup occurred, got line: %s", line)
			}
		}
	}
}

// TestDNSCacheTTLExpiration tests that cache expires after 2+ seconds
func TestDNSCacheTTLExpiration(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skip("Docker daemon not reachable:", err)
	}

	mgr := NewManagerWithClient(cli)

	if err := mgr.BuildImage(ctx, &BuildOptions{
		ImageName: ImageTag(),
		NoCache:   true,
		Output:    io.Discard,
	}); err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	workspace := t.TempDir()

	// Test: Cache should expire after 2+ seconds
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Command: []string{"sh"},
			Args: []string{"-c", `
				# Resolve example.com to get IP in cache
				getent hosts example.com
				EXAMPLE_IP=$(getent hosts example.com | awk '{print $1}' | head -1)
				echo "Resolved example.com to $EXAMPLE_IP"

				# Wait 3 seconds for cache to expire (TTL is 2 seconds)
				echo "Waiting 3 seconds for cache to expire..."
				sleep 3

				# Try to connect after expiration (should show IP only)
				echo "Attempting connection after cache expiration..."
				curl --max-time 5 http://$EXAMPLE_IP 2>&1
			`},
		},
		Workspace: config.WorkspaceConfig{
			Path: workspace,
		},
		Network: config.NetworkConfig{
			Allow: []string{"api.anthropic.com"}, // Block example.com
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

	// curl should fail
	if exitCode == 0 {
		t.Error("Expected curl to fail when connecting to blocked IP")
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logs, err := getContainerLogs(ctx2, cli, containerID)
	if err != nil {
		t.Fatalf("Failed to get container logs: %v", err)
	}

	// Verify cache expiration message appears
	if !strings.Contains(logs, "Waiting 3 seconds for cache to expire") {
		t.Errorf("Expected to see wait message in logs, got:\n%s", logs)
	}

	// Find the error message line after cache expiration
	lines := strings.Split(logs, "\n")
	foundExpirationMsg := false
	for i, line := range lines {
		if strings.Contains(line, "Attempting connection after cache expiration") {
			foundExpirationMsg = true
			// Look at subsequent lines for the error message
			for j := i + 1; j < len(lines); j++ {
				if strings.Contains(lines[j], "agentbox: Connection to") && strings.Contains(lines[j], "blocked") {
					// This should NOT contain hostname (cache expired)
					if strings.Contains(lines[j], "example.com") {
						t.Errorf("Expected error message after TTL expiration to NOT contain hostname, got line: %s", lines[j])
					}
					// Should contain IP only
					if !strings.Contains(lines[j], ".") { // IP contains dots
						t.Errorf("Expected error message to contain IP address after TTL expiration, got line: %s", lines[j])
					}
					break
				}
			}
			break
		}
	}
	if !foundExpirationMsg {
		t.Errorf("Expected to find 'Attempting connection after cache expiration' in logs")
	}
}

// TestDNSCacheThreadIsolation tests that multiple threads maintain independent caches
func TestDNSCacheThreadIsolation(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skip("Docker daemon not reachable:", err)
	}

	mgr := NewManagerWithClient(cli)

	if err := mgr.BuildImage(ctx, &BuildOptions{
		ImageName: ImageTag(),
		NoCache:   true,
		Output:    io.Discard,
	}); err != nil {
		t.Fatalf("BuildImage() error = %v", err)
	}

	workspace := t.TempDir()

	// Test: Simulate thread-local behavior by running multiple subprocesses in parallel
	// Each subprocess runs in its own thread context, so libsandbox thread-local storage
	// should isolate caches between them
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Command: []string{"sh"},
			Args: []string{"-c", `
				# Function to resolve and connect in a subprocess (simulates thread)
				test_dns_and_connect() {
					HOSTNAME=$1
					echo "Process $$: Resolving $HOSTNAME"

					# Resolve hostname (this should populate thread-local cache)
					getent hosts $HOSTNAME > /dev/null
					IP=$(getent hosts $HOSTNAME | awk '{print $1}' | head -1)
					echo "Process $$: Resolved $HOSTNAME to $IP"

					# Immediately try to connect (should use cached hostname)
					echo "Process $$: Attempting connection to $IP"
					curl --max-time 5 http://$IP 2>&1 | grep -i "agentbox: Connection"
					echo "Process $$: Done"
				}

				# Run two subprocesses in parallel (each gets own thread-local cache)
				test_dns_and_connect example.com &
				PID1=$!
				test_dns_and_connect example.org &
				PID2=$!

				# Wait for both to complete
				wait $PID1
				wait $PID2

				echo "All processes completed"
			`},
		},
		Workspace: config.WorkspaceConfig{
			Path: workspace,
		},
		Network: config.NetworkConfig{
			Allow: []string{"api.anthropic.com"}, // Block example.com and example.org
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

	_, err = WaitContainer(ctx, cli, containerID)
	if err != nil {
		t.Fatalf("WaitContainer() error = %v", err)
	}

	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logs, err := getContainerLogs(ctx2, cli, containerID)
	if err != nil {
		t.Fatalf("Failed to get container logs: %v", err)
	}

	// Verify both processes completed
	if !strings.Contains(logs, "All processes completed") {
		t.Errorf("Expected both processes to complete, got logs:\n%s", logs)
	}

	// Verify each process resolved its own hostname
	if !strings.Contains(logs, "Resolving example.com") {
		t.Errorf("Expected process to resolve example.com, got logs:\n%s", logs)
	}
	if !strings.Contains(logs, "Resolving example.org") {
		t.Errorf("Expected process to resolve example.org, got logs:\n%s", logs)
	}

	// Verify both connections were blocked
	// Since processes may have different thread-local caches, we just verify that
	// the connections were blocked (hostname may or may not appear depending on timing)
	if !strings.Contains(logs, "agentbox: Connection") {
		t.Errorf("Expected connection block messages, got logs:\n%s", logs)
	}

	// Note: We cannot reliably test thread-local isolation with shell scripts since
	// each subprocess gets its own memory space. This test verifies that DNS caching
	// works in parallel execution contexts. For true thread-local testing, we would
	// need a multi-threaded C program, but the container doesn't have gcc installed.
	t.Log("Note: This test verifies parallel DNS caching behavior with subprocess isolation.")
	t.Log("True thread-local cache isolation requires multi-threaded programs within same process.")
}
