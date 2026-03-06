package container

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Cleanup ensures a container is stopped and removed.
// This is a best-effort cleanup for edge cases where AutoRemove doesn't work,
// such as when the container fails to start or when cleanup is needed on error paths.
//
// The function silently ignores errors (container already stopped, already removed, etc.)
// since the goal is to ensure no orphan containers are left behind.
func Cleanup(ctx context.Context, cli *client.Client, containerID string) {
	timeout := 5 * time.Second
	timeoutSec := int(timeout.Seconds())

	// Stop the container (may already be stopped)
	_ = cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeoutSec,
	})

	// Remove the container (may already be removed by AutoRemove)
	_ = cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}
