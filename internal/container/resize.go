package container

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/moby/term"
)

// HandleResize listens for SIGWINCH signals (terminal resize events) and resizes
// the container's PTY to match the host terminal dimensions.
// This function should only be called when TTY is allocated.
//
// It performs an initial resize to sync dimensions, then continues listening
// for resize events in a background goroutine.
//
// The function returns a cleanup function that should be called to stop
// listening for resize signals (typically via defer).
func HandleResize(ctx context.Context, cli *client.Client, containerID string) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	// Initial resize to sync dimensions
	resizeContainer(ctx, cli, containerID)

	go func() {
		for range sigCh {
			resizeContainer(ctx, cli, containerID)
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}
}

// resizeContainer resizes the container's TTY to match the current terminal size.
// If the terminal size cannot be determined (e.g., not a terminal), this is a no-op.
func resizeContainer(ctx context.Context, cli *client.Client, containerID string) {
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		return
	}

	ws, err := term.GetWinsize(fd)
	if err != nil {
		return
	}

	_ = cli.ContainerResize(ctx, containerID, container.ResizeOptions{
		Width:  uint(ws.Width),
		Height: uint(ws.Height),
	})
}
