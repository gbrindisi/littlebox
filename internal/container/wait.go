package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// WaitContainer waits for a container to exit and returns its exit code.
// This is a standalone function for CLI usage, complementing AttachContainer.
//
// The function blocks until the container stops running. It returns the
// container's exit code (0 for success, non-zero for failure) and any
// error that occurred during the wait operation.
//
// Use this function to capture the exit code for returning to the shell,
// allowing agentbox to exit with the same code as the agent inside.
func WaitContainer(ctx context.Context, cli *client.Client, containerID string) (int, error) {
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		return 1, fmt.Errorf("error waiting for container: %w", err)
	case status := <-statusCh:
		return int(status.StatusCode), nil
	case <-ctx.Done():
		return 1, ctx.Err()
	}
}
