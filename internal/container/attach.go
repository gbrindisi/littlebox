package container

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gbrindisi/littlebox/internal/output"
	"github.com/moby/term"
)

// filteringWriter wraps an io.Writer and filters out lines matching certain prefixes in quiet mode.
// In debug mode, all lines pass through unchanged.
type filteringWriter struct {
	w          io.Writer
	verbosity  output.Verbosity
	buf        []byte
	setupDone  bool
	statusDone func() // callback to print "done" when setup completes
}

func newFilteringWriter(w io.Writer, verbosity output.Verbosity, statusDone func()) *filteringWriter {
	return &filteringWriter{
		w:          w,
		verbosity:  verbosity,
		statusDone: statusDone,
	}
}

func (fw *filteringWriter) Write(p []byte) (n int, err error) {
	// In debug mode, pass through all output unchanged
	if fw.verbosity == output.Debug {
		return fw.w.Write(p)
	}

	// After setup is done, pass through all output immediately without buffering.
	// This is critical for interactive TTY sessions (shell mode) where character-by-character
	// echo from bash would otherwise get stuck in the buffer waiting for newlines.
	if fw.setupDone {
		return fw.w.Write(p)
	}

	// In quiet mode during setup, filter out entrypoint and firewall logs
	fw.buf = append(fw.buf, p...)

	// Process complete lines
	for {
		idx := strings.IndexByte(string(fw.buf), '\n')
		if idx == -1 {
			break
		}

		line := string(fw.buf[:idx+1])
		fw.buf = fw.buf[idx+1:]

		// Filter out [entrypoint] and [firewall] lines
		if strings.Contains(line, "[entrypoint]") || strings.Contains(line, "[firewall]") {
			// These are setup logs, skip them

			// Check if this is the last entrypoint line (dropping privileges)
			if !fw.setupDone && strings.Contains(line, "Dropping privileges") {
				fw.setupDone = true
				if fw.statusDone != nil {
					fw.statusDone()
				}
				// Flush any remaining buffered data now that setup is done
				if len(fw.buf) > 0 {
					_, _ = fw.w.Write(fw.buf)
					fw.buf = nil
				}
			}
			continue
		}

		// Pass through other lines (agent output)
		if _, err := fw.w.Write([]byte(line)); err != nil {
			return len(p), err
		}
	}

	return len(p), nil
}

// AttachContainer attaches stdin/stdout/stderr to a container and starts it.
// The tty parameter determines how output is handled:
// - With TTY: stdout and stderr are multiplexed together
// - Without TTY: stdout and stderr are demultiplexed using Docker's stdcopy
// The verbosity parameter controls output filtering:
// - Quiet: filters out entrypoint/firewall logs, shows only agent output
// - Debug: shows all output including setup logs
//
// This function blocks until the container exits or the context is cancelled.
// It returns nil on successful completion, or an error if attachment or I/O fails.
func AttachContainer(ctx context.Context, cli *client.Client, containerID string, tty bool, verbosity output.Verbosity, statusDone func()) error {
	resp, err := cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return err
	}

	// Set terminal to raw mode for TTY containers so TUI apps receive keystrokes
	if tty {
		oldState, rawErr := term.SetRawTerminal(os.Stdin.Fd())
		if rawErr != nil {
			// Log but continue - raw mode is nice to have but not critical
			// This can fail if stdin is not a terminal (e.g., piped input)
			_, _ = os.Stderr.WriteString("[agentbox] warning: could not set raw terminal: " + rawErr.Error() + "\n")
		} else {
			defer func() { _ = term.RestoreTerminal(os.Stdin.Fd(), oldState) }()
		}
	}

	outputDone := make(chan error, 1)
	stdinDone := make(chan struct{})

	// Copy stdin to container
	go func() {
		defer close(stdinDone)
		_, _ = io.Copy(resp.Conn, os.Stdin)
		// Signal EOF to container stdin after input is complete
		_ = resp.CloseWrite()
	}()

	// Create filtering writer for stdout in quiet mode
	stdoutWriter := newFilteringWriter(os.Stdout, verbosity, statusDone)

	// Copy container output to stdout/stderr
	go func() {
		var err error
		if tty {
			// With TTY, output is raw bytes - use filtering writer
			_, err = io.Copy(stdoutWriter, resp.Reader)
		} else {
			// Without TTY, demultiplex stdout and stderr
			// stdout goes through filter, stderr goes directly
			_, err = stdcopy.StdCopy(stdoutWriter, os.Stderr, resp.Reader)
		}
		outputDone <- err
	}()

	// Wait for either output to complete (container exit) or stdin to close.
	// Use select to avoid blocking forever on stdin when container exits first.
	select {
	case err = <-outputDone:
		// Container exited, return immediately without waiting for stdin
	case <-stdinDone:
		// Stdin closed, wait for output to finish
		err = <-outputDone
	}

	return err
}
