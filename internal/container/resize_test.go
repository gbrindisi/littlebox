package container

import (
	"syscall"
	"testing"
)

func TestSIGWINCHSignal(t *testing.T) {
	// Verify SIGWINCH is the expected signal constant
	// This ensures we're using the correct signal for terminal resize
	if syscall.SIGWINCH != syscall.Signal(0x1c) {
		t.Errorf("SIGWINCH has unexpected value: got %v, expected 0x1c (28)", syscall.SIGWINCH)
	}
}
