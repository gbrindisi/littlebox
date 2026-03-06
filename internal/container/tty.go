package container

import (
	"os"

	"github.com/moby/term"
)

// TTYMode controls TTY allocation behavior.
type TTYMode int

const (
	// TTYAuto automatically detects if stdin is a terminal.
	TTYAuto TTYMode = iota
	// TTYForce forces TTY allocation even when stdin is not a terminal.
	TTYForce
	// TTYNone disables TTY allocation even when stdin is a terminal.
	TTYNone
)

// DetectTTY determines whether to allocate a TTY based on the mode.
// With TTYAuto, it checks if stdin is a terminal.
// TTYForce always returns true, TTYNone always returns false.
func DetectTTY(mode TTYMode) bool {
	switch mode {
	case TTYForce:
		return true
	case TTYNone:
		return false
	default:
		return term.IsTerminal(os.Stdin.Fd())
	}
}
