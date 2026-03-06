package container

import (
	"testing"
)

func TestDetectTTY_ForceMode(t *testing.T) {
	result := DetectTTY(TTYForce)
	if !result {
		t.Error("TTYForce should always return true")
	}
}

func TestDetectTTY_NoneMode(t *testing.T) {
	result := DetectTTY(TTYNone)
	if result {
		t.Error("TTYNone should always return false")
	}
}

func TestDetectTTY_AutoMode(t *testing.T) {
	// In test environment, stdin is typically not a terminal
	// This test verifies that DetectTTY doesn't panic and returns a bool
	result := DetectTTY(TTYAuto)
	// Result depends on whether stdin is a terminal
	// In CI/test environments this is typically false
	_ = result // Just verify it doesn't panic
}

func TestTTYModeConstants(t *testing.T) {
	// Verify constants have expected values (iota starts at 0)
	if TTYAuto != 0 {
		t.Errorf("TTYAuto should be 0, got %d", TTYAuto)
	}
	if TTYForce != 1 {
		t.Errorf("TTYForce should be 1, got %d", TTYForce)
	}
	if TTYNone != 2 {
		t.Errorf("TTYNone should be 2, got %d", TTYNone)
	}
}
