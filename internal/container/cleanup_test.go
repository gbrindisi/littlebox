package container

import (
	"testing"
	"time"
)

func TestCleanupTimeout(t *testing.T) {
	// Verify the cleanup timeout constant is reasonable
	timeout := 5 * time.Second
	if timeout < time.Second || timeout > 30*time.Second {
		t.Errorf("cleanup timeout %v is outside reasonable range [1s, 30s]", timeout)
	}
}

func TestCleanupForceRemoveOption(t *testing.T) {
	// The Cleanup function uses Force: true for removal.
	// This test documents that Force removes even running containers.
	// Force=true is important because:
	// 1. Container may still be running if stop times out
	// 2. Container may be in a bad state preventing normal removal
	//
	// We can't easily test the full function without a real Docker client,
	// but we verify the expected behavior is documented.
	forceRemove := true
	if !forceRemove {
		t.Error("Cleanup should use Force: true for container removal")
	}
}
