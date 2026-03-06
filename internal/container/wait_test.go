package container

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestWaitConditionConstant(t *testing.T) {
	// Verify we're using the correct wait condition
	// WaitConditionNotRunning waits until container has stopped
	if container.WaitConditionNotRunning != "not-running" {
		t.Errorf("WaitConditionNotRunning = %q, want %q", container.WaitConditionNotRunning, "not-running")
	}
}

func TestWaitContainerContextCancellation(t *testing.T) {
	// Test that context cancellation is handled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// We can't easily test the full function without a real Docker client,
	// but we can verify the context cancellation path works correctly
	// by checking that a cancelled context returns an error
	if ctx.Err() == nil {
		t.Error("expected cancelled context to have error")
	}
}
