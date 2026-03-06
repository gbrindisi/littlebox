package output

import (
	"fmt"
	"io"
)

// Verbosity controls the level of output detail
type Verbosity int

const (
	// Quiet shows minimal formatted output (default)
	Quiet Verbosity = iota
	// Debug shows verbose output with all logs
	Debug
)

// BulletPrint prints a message with a bullet prefix in quiet mode.
// In debug mode, it prints nothing (caller should use regular output).
// Format: "● message"
func BulletPrint(w io.Writer, verbosity Verbosity, msg string) {
	if verbosity == Quiet {
		_, _ = fmt.Fprintf(w, "● %s\r\n", msg)
	}
}

// StatusWriter handles the "message... done" pattern for long-running operations.
// In quiet mode: prints "● message..." when started, prints " done" when finished.
// In debug mode: does nothing (caller should use regular output).
type StatusWriter struct {
	w         io.Writer
	verbosity Verbosity
	started   bool
}

// NewStatusWriter creates a new StatusWriter for the given verbosity level
func NewStatusWriter(w io.Writer, verbosity Verbosity) *StatusWriter {
	return &StatusWriter{
		w:         w,
		verbosity: verbosity,
	}
}

// Start prints the initial message with "..." suffix in quiet mode
func (s *StatusWriter) Start(msg string) {
	if s.verbosity == Quiet {
		_, _ = fmt.Fprintf(s.w, "● %s...", msg)
		s.started = true
	}
}

// Done prints " done" to complete the status line in quiet mode
func (s *StatusWriter) Done() {
	if s.verbosity == Quiet && s.started {
		_, _ = fmt.Fprint(s.w, " done\r\n")
		s.started = false
	}
}

// Failed prints " failed" to indicate operation failure in quiet mode
func (s *StatusWriter) Failed() {
	if s.verbosity == Quiet && s.started {
		_, _ = fmt.Fprint(s.w, " failed\r\n")
		s.started = false
	}
}

// Writer returns an io.Writer that can be used for operations in the given verbosity mode.
// In quiet mode: returns io.Discard to suppress output
// In debug mode: returns the provided writer for full output
func Writer(w io.Writer, verbosity Verbosity) io.Writer {
	if verbosity == Quiet {
		return io.Discard
	}
	return w
}
