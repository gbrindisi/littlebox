package container

import (
	"syscall"
	"testing"
)

func TestSignalStrings(t *testing.T) {
	// Test that the signal constants we handle are correct
	tests := []struct {
		name     string
		signal   syscall.Signal
		expected string
	}{
		{
			name:     "SIGINT is signal 2",
			signal:   syscall.SIGINT,
			expected: "interrupt",
		},
		{
			name:     "SIGTERM is signal 15",
			signal:   syscall.SIGTERM,
			expected: "terminated",
		},
		{
			name:     "SIGQUIT is signal 3",
			signal:   syscall.SIGQUIT,
			expected: "quit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.signal.String()
			if got != tt.expected {
				t.Errorf("signal.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSignalNumbers(t *testing.T) {
	// Verify signal numbers match Unix conventions
	if syscall.SIGINT != 2 {
		t.Errorf("SIGINT = %d, want 2", syscall.SIGINT)
	}
	if syscall.SIGTERM != 15 {
		t.Errorf("SIGTERM = %d, want 15", syscall.SIGTERM)
	}
	if syscall.SIGQUIT != 3 {
		t.Errorf("SIGQUIT = %d, want 3", syscall.SIGQUIT)
	}
}

func TestSignalAction(t *testing.T) {
	tests := []struct {
		sig      syscall.Signal
		wantKill string
		wantStop bool
	}{
		{syscall.SIGINT, "SIGINT", false},
		{syscall.SIGTERM, "SIGTERM", false},
		{syscall.SIGQUIT, "SIGQUIT", false},
		{syscall.SIGHUP, "SIGTERM", true},
		{syscall.SIGUSR1, "", false},
	}
	for _, tt := range tests {
		gotKill, gotStop := signalAction(tt.sig)
		if gotKill != tt.wantKill || gotStop != tt.wantStop {
			t.Errorf("signalAction(%v) = (%q, %v), want (%q, %v)", tt.sig, gotKill, gotStop, tt.wantKill, tt.wantStop)
		}
	}
}

func TestForwardedSignalsIncludesSIGHUP(t *testing.T) {
	for _, s := range forwardedSignals {
		if s == syscall.SIGHUP {
			return
		}
	}
	t.Fatal("SIGHUP must be intercepted so deferred Cleanup runs instead of Go's default exit")
}
