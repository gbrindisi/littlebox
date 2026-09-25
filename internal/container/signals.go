package container

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// hangupStopTimeout is how long the container gets to exit after SIGTERM
// (sent on SIGHUP) before Docker kills it.
const hangupStopTimeout = 10 * time.Second

// forwardedSignals is the set of host signals littlebox intercepts.
var forwardedSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}

// signalAction describes how a host signal is handled.
// If stop is true, the container is stopped (SIGTERM, then SIGKILL after a timeout).
// Otherwise killSignal is forwarded verbatim via ContainerKill. Empty killSignal
// with stop=false means the signal is ignored.
func signalAction(sig os.Signal) (killSignal string, stop bool) {
	switch sig {
	case syscall.SIGINT:
		return "SIGINT", false
	case syscall.SIGTERM:
		return "SIGTERM", false
	case syscall.SIGQUIT:
		return "SIGQUIT", false
	case syscall.SIGHUP:
		// Controlling terminal went away (e.g. tmux window killed). Nobody is
		// left to interact with the agent, so stop the container. Intercepting
		// SIGHUP also prevents Go's default "exit immediately" behavior, so the
		// run path's deferred Cleanup still executes.
		return "SIGTERM", true
	default:
		return "", false
	}
}

// ForwardSignals forwards host signals to the container.
// SIGINT, SIGTERM and SIGQUIT are forwarded via ContainerKill. SIGHUP is
// treated like SIGTERM but via ContainerStop, which escalates to SIGKILL after
// hangupStopTimeout so the container is guaranteed to exit; WaitContainer then
// returns and the caller's deferred Cleanup removes the container.
//
// The returned function stops signal forwarding (typically via defer).
func ForwardSignals(ctx context.Context, cli *client.Client, containerID string) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, forwardedSignals...)

	go func() {
		for sig := range sigCh {
			sigStr, stop := signalAction(sig)
			if stop {
				timeoutSec := int(hangupStopTimeout.Seconds())
				_ = cli.ContainerStop(ctx, containerID, container.StopOptions{Signal: sigStr, Timeout: &timeoutSec})
				continue
			}
			if sigStr == "" {
				continue
			}
			// Best-effort: container may already have exited ("No such container").
			_ = cli.ContainerKill(ctx, containerID, sigStr)
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}
}
