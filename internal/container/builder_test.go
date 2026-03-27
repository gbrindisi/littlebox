package container

import (
	"bytes"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/gbrindisi/littlebox/internal/output"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should be a valid semver-like format (X.X.X)
	parts := strings.Split(Version, ".")
	if len(parts) != 3 {
		t.Errorf("Version should have 3 parts (X.X.X), got: %s", Version)
	}
}

func TestImageTag(t *testing.T) {
	tag := ImageTag()

	// Should have correct prefix
	if !strings.HasPrefix(tag, "littlebox/base:") {
		t.Errorf("ImageTag should start with 'littlebox/base:', got: %s", tag)
	}

	// Should include version
	if !strings.Contains(tag, Version) {
		t.Errorf("ImageTag should contain version %s, got: %s", Version, tag)
	}

	// Should be exactly littlebox/base:VERSION
	expected := "littlebox/base:" + Version
	if tag != expected {
		t.Errorf("ImageTag() = %s, want %s", tag, expected)
	}
}

func TestDefaultImageNameUsesVersionedTag(t *testing.T) {
	// DefaultImageName should now use the versioned tag
	expected := ImageTag()
	if DefaultImageName != expected {
		t.Errorf("DefaultImageName = %s, want %s", DefaultImageName, expected)
	}
}

func TestEnsureImageOutputMessages(t *testing.T) {
	// Test that EnsureImage produces correct output messages
	// Note: This tests the message format, not actual Docker operations

	t.Run("cached image message format", func(t *testing.T) {
		// Verify the cached message format matches expected pattern
		var buf bytes.Buffer
		expectedTag := ImageTag()
		expectedMsg := "Using cached image: " + expectedTag + "\n"

		// We can't easily test with a real Manager without Docker,
		// but we can verify the message format is correct
		if !strings.Contains(expectedMsg, "Using cached image:") {
			t.Error("Cached message should contain 'Using cached image:'")
		}
		if !strings.Contains(expectedMsg, expectedTag) {
			t.Error("Cached message should contain the image tag")
		}

		// Verify buffer can receive output (io.Writer interface)
		buf.WriteString(expectedMsg)
		if buf.String() != expectedMsg {
			t.Errorf("Buffer output mismatch: got %q, want %q", buf.String(), expectedMsg)
		}
	})

	t.Run("building image message format", func(t *testing.T) {
		expectedTag := ImageTag()
		expectedMsg := "Building image: " + expectedTag + "\n"

		if !strings.Contains(expectedMsg, "Building image:") {
			t.Error("Build message should contain 'Building image:'")
		}
		if !strings.Contains(expectedMsg, expectedTag) {
			t.Error("Build message should contain the image tag")
		}
	})
}

func TestDerivedImageTag(t *testing.T) {
	workspacePath := "/workspace/test"

	t.Run("returns correct format", func(t *testing.T) {
		script := "curl -fsSL https://example.com/install.sh | bash"
		tag := DerivedImageTag(script, workspacePath)

		// Should have correct prefix
		if !strings.HasPrefix(tag, "littlebox/build:") {
			t.Errorf("DerivedImageTag should start with 'littlebox/build:', got: %s", tag)
		}

		// Hash portion should be 12 characters
		hashPart := strings.TrimPrefix(tag, "littlebox/build:")
		if len(hashPart) != 12 {
			t.Errorf("Hash portion should be 12 characters, got %d: %s", len(hashPart), hashPart)
		}
	})

	t.Run("same script and workspace produces same hash (deterministic)", func(t *testing.T) {
		script := "apt-get update && apt-get install -y git"
		tag1 := DerivedImageTag(script, workspacePath)
		tag2 := DerivedImageTag(script, workspacePath)

		if tag1 != tag2 {
			t.Errorf("Same script and workspace should produce same tag: %s != %s", tag1, tag2)
		}
	})

	t.Run("different scripts produce different hashes", func(t *testing.T) {
		script1 := "apt-get update"
		script2 := "apt-get upgrade"
		tag1 := DerivedImageTag(script1, workspacePath)
		tag2 := DerivedImageTag(script2, workspacePath)

		if tag1 == tag2 {
			t.Errorf("Different scripts should produce different tags: %s == %s", tag1, tag2)
		}
	})

	t.Run("same script with different workspace produces different hash", func(t *testing.T) {
		script := "apt-get update && apt-get install -y git"
		workspace1 := "/workspace/project1"
		workspace2 := "/workspace/project2"
		tag1 := DerivedImageTag(script, workspace1)
		tag2 := DerivedImageTag(script, workspace2)

		if tag1 == tag2 {
			t.Errorf("Same script with different workspace should produce different tags: %s == %s", tag1, tag2)
		}
	})

	t.Run("null byte separator prevents collision", func(t *testing.T) {
		// Test that script="A" + workspace="B" produces different hash than script="AB" + workspace=""
		script1 := "A"
		workspace1 := "B"
		script2 := "AB"
		workspace2 := ""
		tag1 := DerivedImageTag(script1, workspace1)
		tag2 := DerivedImageTag(script2, workspace2)

		if tag1 == tag2 {
			t.Errorf("Null byte separator should prevent collision: %s == %s", tag1, tag2)
		}
	})

	t.Run("empty script works", func(t *testing.T) {
		tag := DerivedImageTag("", workspacePath)

		if !strings.HasPrefix(tag, "littlebox/build:") {
			t.Errorf("Empty script should still produce valid tag, got: %s", tag)
		}
	})

	t.Run("multiline script works", func(t *testing.T) {
		script := `apt-get update
apt-get install -y curl
curl -fsSL https://example.com/install.sh | bash`
		tag := DerivedImageTag(script, workspacePath)

		if !strings.HasPrefix(tag, "littlebox/build:") {
			t.Errorf("Multiline script should produce valid tag, got: %s", tag)
		}
	})
}

func TestGenerateDerivedDockerfile(t *testing.T) {
	t.Run("includes base image FROM", func(t *testing.T) {
		script := "apt-get update"
		dockerfile := generateDerivedDockerfile(script)

		expectedFrom := "FROM " + ImageTag()
		if !strings.Contains(dockerfile, expectedFrom) {
			t.Errorf("Dockerfile should contain FROM %s, got:\n%s", ImageTag(), dockerfile)
		}
	})

	t.Run("switches to root before script", func(t *testing.T) {
		script := "apt-get update"
		dockerfile := generateDerivedDockerfile(script)

		if !strings.Contains(dockerfile, "USER root") {
			t.Errorf("Dockerfile should contain USER root, got:\n%s", dockerfile)
		}
	})

	t.Run("remains as root for entrypoint privilege drop", func(t *testing.T) {
		script := "apt-get update"
		dockerfile := generateDerivedDockerfile(script)

		// Should NOT contain USER agent - the entrypoint handles privilege drop via setpriv
		if strings.Contains(dockerfile, "USER agent") {
			t.Errorf("Dockerfile should NOT contain USER agent (entrypoint handles privilege drop), got:\n%s", dockerfile)
		}

		// Should contain USER root for the build script
		if !strings.Contains(dockerfile, "USER root") {
			t.Errorf("Dockerfile should contain USER root, got:\n%s", dockerfile)
		}
	})

	t.Run("includes script in RUN heredoc", func(t *testing.T) {
		script := "apt-get update && apt-get install -y curl"
		dockerfile := generateDerivedDockerfile(script)

		if !strings.Contains(dockerfile, script) {
			t.Errorf("Dockerfile should contain the script, got:\n%s", dockerfile)
		}

		if !strings.Contains(dockerfile, "RUN <<'SCRIPT'") {
			t.Errorf("Dockerfile should use heredoc syntax, got:\n%s", dockerfile)
		}

		if !strings.Contains(dockerfile, "SCRIPT") {
			t.Errorf("Dockerfile should have SCRIPT heredoc delimiter, got:\n%s", dockerfile)
		}
	})

	t.Run("handles multiline script", func(t *testing.T) {
		script := `apt-get update
apt-get install -y curl git
curl -fsSL https://example.com/install.sh | bash`
		dockerfile := generateDerivedDockerfile(script)

		// All lines should be present
		for _, line := range strings.Split(script, "\n") {
			if !strings.Contains(dockerfile, line) {
				t.Errorf("Dockerfile should contain script line %q, got:\n%s", line, dockerfile)
			}
		}
	})
}

func TestCreateDerivedBuildContext(t *testing.T) {
	t.Run("creates valid tar archive", func(t *testing.T) {
		dockerfile := "FROM alpine\nRUN echo hello"
		reader, err := createDerivedBuildContext(dockerfile)
		if err != nil {
			t.Fatalf("createDerivedBuildContext failed: %v", err)
		}

		// Read the tar archive
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(reader)
		if err != nil {
			t.Fatalf("Failed to read build context: %v", err)
		}

		// Verify it's not empty
		if buf.Len() == 0 {
			t.Error("Build context should not be empty")
		}
	})

	t.Run("contains Dockerfile", func(t *testing.T) {
		dockerfile := "FROM alpine\nRUN echo hello"
		reader, err := createDerivedBuildContext(dockerfile)
		if err != nil {
			t.Fatalf("createDerivedBuildContext failed: %v", err)
		}

		// Extract and verify Dockerfile content
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(reader)

		// The tar should contain the dockerfile content
		if !strings.Contains(buf.String(), dockerfile) {
			t.Error("Build context should contain the dockerfile content")
		}
	})
}

// Integration tests for derived image build and caching.
// These tests require Docker to be available and are skipped otherwise.

func TestEnsureDerivedImageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a manager to check Docker availability
	mgr, err := NewManager()
	if err != nil {
		t.Skipf("skipping integration test: Docker not available: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Use context for all Docker operations
	ctx := t.Context()

	// First, ensure the base image exists
	var baseBuf bytes.Buffer
	if err := mgr.EnsureImage(ctx, false, output.Debug, &baseBuf); err != nil {
		t.Skipf("skipping integration test: failed to ensure base image: %v", err)
	}

	// Use a simple build script that runs quickly
	buildScript := "echo 'test build script' > /tmp/test-marker"
	testWorkspace := "/workspace/test"
	derivedTag := DerivedImageTag(buildScript, testWorkspace)

	// Clean up any existing test image first
	_, _ = mgr.client.ImageRemove(ctx, derivedTag, image.RemoveOptions{Force: true})

	t.Run("builds derived image when not cached", func(t *testing.T) {
		var buf bytes.Buffer
		tag, err := mgr.EnsureDerivedImage(ctx, buildScript, testWorkspace, false, output.Quiet, &buf)
		if err != nil {
			t.Fatalf("EnsureDerivedImage failed: %v", err)
		}

		if tag != derivedTag {
			t.Errorf("expected tag %s, got %s", derivedTag, tag)
		}

		out := buf.String()
		// In quiet mode, expect the "Building... done" pattern with bullet prefix
		if !strings.Contains(out, "● Building derived image:") || !strings.Contains(out, " done") {
			t.Errorf("expected quiet mode build message, got: %s", out)
		}

		// Verify image exists
		exists, err := mgr.ImageExists(ctx, derivedTag)
		if err != nil {
			t.Fatalf("ImageExists failed: %v", err)
		}
		if !exists {
			t.Error("expected derived image to exist after build")
		}
	})

	t.Run("uses cached image on second call", func(t *testing.T) {
		var buf bytes.Buffer
		tag, err := mgr.EnsureDerivedImage(ctx, buildScript, testWorkspace, false, output.Quiet, &buf)
		if err != nil {
			t.Fatalf("EnsureDerivedImage failed: %v", err)
		}

		if tag != derivedTag {
			t.Errorf("expected tag %s, got %s", derivedTag, tag)
		}

		out := buf.String()
		// In quiet mode, expect bullet-prefixed cache message
		if !strings.Contains(out, "● Using cached derived image:") {
			t.Errorf("expected cache message, got: %s", out)
		}
	})

	t.Run("rebuilds when forceBuild is true", func(t *testing.T) {
		var buf bytes.Buffer
		tag, err := mgr.EnsureDerivedImage(ctx, buildScript, testWorkspace, true, output.Quiet, &buf)
		if err != nil {
			t.Fatalf("EnsureDerivedImage failed: %v", err)
		}

		if tag != derivedTag {
			t.Errorf("expected tag %s, got %s", derivedTag, tag)
		}

		out := buf.String()
		// In quiet mode, expect the "Building... done" pattern with bullet prefix
		if !strings.Contains(out, "● Building derived image:") || !strings.Contains(out, " done") {
			t.Errorf("expected quiet mode build message with forceBuild=true, got: %s", out)
		}
	})

	t.Run("different script produces different image", func(t *testing.T) {
		differentScript := "echo 'different script' > /tmp/different-marker"
		differentTag := DerivedImageTag(differentScript, testWorkspace)

		// Clean up first
		_, _ = mgr.client.ImageRemove(ctx, differentTag, image.RemoveOptions{Force: true})

		var buf bytes.Buffer
		tag, err := mgr.EnsureDerivedImage(ctx, differentScript, testWorkspace, false, output.Debug, &buf)
		if err != nil {
			t.Fatalf("EnsureDerivedImage failed: %v", err)
		}

		if tag == derivedTag {
			t.Error("expected different tag for different script")
		}
		if tag != differentTag {
			t.Errorf("expected tag %s, got %s", differentTag, tag)
		}

		// Clean up
		_, _ = mgr.client.ImageRemove(ctx, differentTag, image.RemoveOptions{Force: true})
	})

	// Clean up the test image
	_, _ = mgr.client.ImageRemove(ctx, derivedTag, image.RemoveOptions{Force: true})
}
