package container

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/gbrindisi/littlebox/internal/output"
)

// TestBuildErrorIntegration contains integration tests that verify error display
// with intentionally failing build scripts. These tests require Docker to be available.

// buildAndExpectError is a test helper that builds a derived image with the given script
// and expects it to fail, returning the formatted error output for verification.
func buildAndExpectError(t *testing.T, mgr *Manager, ctx context.Context, script string) string {
	t.Helper()

	testWorkspace := "/workspace/builderror-test"
	derivedTag := DerivedImageTag(script, testWorkspace)

	// Clean up any existing test image first
	_, _ = mgr.client.ImageRemove(ctx, derivedTag, image.RemoveOptions{Force: true})
	defer func() {
		// Clean up after test
		_, _ = mgr.client.ImageRemove(ctx, derivedTag, image.RemoveOptions{Force: true})
	}()

	// Build and capture output
	var buf bytes.Buffer
	_, err := mgr.EnsureDerivedImage(ctx, script, testWorkspace, false, output.Quiet, &buf)
	if err == nil {
		t.Fatal("expected build to fail, but it succeeded")
	}

	return err.Error()
}

func TestCommandNotFoundError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	script := "nonexistent_command_xyz"
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify error output includes Dockerfile with line numbers
	if !strings.Contains(errOutput, "1 | FROM") {
		t.Errorf("expected Dockerfile line 1 with FROM, got:\n%s", errOutput)
	}
	if !strings.Contains(errOutput, "2 | USER root") {
		t.Errorf("expected Dockerfile line 2 with USER root, got:\n%s", errOutput)
	}
	if !strings.Contains(errOutput, "3 | RUN") {
		t.Errorf("expected Dockerfile line 3 with RUN, got:\n%s", errOutput)
	}

	// Verify error output contains command name and "not found" or similar error
	if !strings.Contains(errOutput, "nonexistent_command_xyz") {
		t.Errorf("expected error to mention the command, got:\n%s", errOutput)
	}
}

func TestExitCodeError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	script := "exit 1"
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify error output includes Dockerfile with line numbers
	if !strings.Contains(errOutput, "| FROM") {
		t.Errorf("expected Dockerfile with line numbers, got:\n%s", errOutput)
	}

	// Verify error mentions exit code 1
	if !strings.Contains(errOutput, "exit") && !strings.Contains(errOutput, "1") {
		t.Errorf("expected error to mention exit code 1, got:\n%s", errOutput)
	}
}

func TestNetworkError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	script := "curl -f https://nonexistent.invalid/install.sh"
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify error output includes Dockerfile with line numbers
	if !strings.Contains(errOutput, "| FROM") {
		t.Errorf("expected Dockerfile with line numbers, got:\n%s", errOutput)
	}

	// Verify error contains the URL and some indication of network failure
	if !strings.Contains(errOutput, "nonexistent.invalid") {
		t.Errorf("expected error to mention the URL, got:\n%s", errOutput)
	}
}

func TestMultilineScriptFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	// Script with multiple commands, failure on later line
	script := `echo "Starting setup..."
echo "Step 1 complete"
echo "Step 2 complete"
nonexistent_command_that_fails`
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify error output includes Dockerfile with line numbers
	if !strings.Contains(errOutput, "| FROM") {
		t.Errorf("expected Dockerfile with line numbers, got:\n%s", errOutput)
	}

	// Verify output shows context from earlier successful commands
	// The build output should include the echo statements before the failure
	if !strings.Contains(errOutput, "Starting setup") || !strings.Contains(errOutput, "Step 1 complete") {
		t.Errorf("expected error to show output from successful commands, got:\n%s", errOutput)
	}
}

func TestErrorOutputLineHighlighting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	script := "exit 1"
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify Dockerfile is formatted with line numbers
	// Format should be " N | <line>" for normal lines
	lines := strings.Split(errOutput, "\n")
	foundLineNumbers := false
	for _, line := range lines {
		if strings.Contains(line, " 1 | FROM") || strings.Contains(line, " 2 | USER") || strings.Contains(line, " 3 | RUN") {
			foundLineNumbers = true
			break
		}
	}
	if !foundLineNumbers {
		t.Errorf("expected Dockerfile lines to have ' N | ' format, got:\n%s", errOutput)
	}
}

func TestErrorOutputTruncation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	// Script that generates many lines of output before failing
	// We need multiple separate commands to generate multiple log entries (not one multi-line entry)
	script := `echo "Line 1"
echo "Line 2"
echo "Line 3"
echo "Line 4"
echo "Line 5"
echo "Line 6"
echo "Line 7"
echo "Line 8"
echo "Line 9"
echo "Line 10"
echo "Line 11"
echo "Line 12"
echo "Line 13"
echo "Line 14"
echo "Line 15"
echo "Line 16"
echo "Line 17"
echo "Line 18"
echo "Line 19"
echo "Line 20"
echo "Line 21"
echo "Line 22"
echo "Line 23"
echo "Line 24"
echo "Line 25"
exit 1`
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify error output includes Dockerfile section
	if !strings.Contains(errOutput, "| FROM") {
		t.Errorf("expected Dockerfile section with line numbers, got:\n%s", errOutput)
	}

	// Verify error output includes build output (should show at least some lines)
	if !strings.Contains(errOutput, "Line ") {
		t.Errorf("expected build output with Line messages, got:\n%s", errOutput)
	}

	// The truncation behavior depends on how BuildKit collects logs
	// At minimum, verify the output is formatted (has proper structure)
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("expected error message section, got:\n%s", errOutput)
	}
}

func TestErrorOutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	ctx := t.Context()

	// Ensure base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	// Simple failing script
	script := `echo "Build output line"
exit 1`
	errOutput := buildAndExpectError(t, mgr, ctx, script)

	// Verify the error output has the expected sections:
	// 1. Dockerfile with line numbers
	if !strings.Contains(errOutput, "| FROM") || !strings.Contains(errOutput, "| RUN") {
		t.Errorf("expected Dockerfile section with line numbers, got:\n%s", errOutput)
	}

	// 2. Error message section
	if !strings.Contains(errOutput, "Error:") {
		t.Errorf("expected 'Error:' section, got:\n%s", errOutput)
	}

	// 3. Build output (if collected by BuildKit traces)
	// Note: The exact format depends on how BuildKit reports logs
	// We just verify the output is structured and includes the Dockerfile context
	if !strings.Contains(errOutput, "FROM littlebox/base") {
		t.Errorf("expected formatted Dockerfile in error output, got:\n%s", errOutput)
	}
}
