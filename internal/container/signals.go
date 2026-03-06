package container

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/docker/docker/client"
)

// ForwardSignals forwards host signals to the container.
// It sets up signal handlers for SIGINT (Ctrl+C), SIGTERM (kill), and SIGQUIT,
// and forwards them to the container via Docker's ContainerKill API.
// This allows the container process to handle cleanup gracefully.
//
// The function returns a cleanup function that should be called to stop
// signal forwarding (typically via defer). This stops the signal notification
// and prevents further signals from being forwarded.
func ForwardSignals(ctx context.Context, cli *client.Client, containerID string) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for sig := range sigCh {
			var sigStr string
			switch sig {
			case syscall.SIGINT:
				sigStr = "SIGINT"
			case syscall.SIGTERM:
				sigStr = "SIGTERM"
			case syscall.SIGQUIT:
				sigStr = "SIGQUIT"
			default:
				continue
			}

			if err := cli.ContainerKill(ctx, containerID, sigStr); err != nil {
				// Best-effort signal forwarding - silently ignore if container already exited
				// "No such container" is expected when the container has finished
				if !strings.Contains(err.Error(), "No such container") {
					// Log unexpected errors for debugging
					_ = err // Silently ignore - this is best-effort
				}
			}
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}
}
