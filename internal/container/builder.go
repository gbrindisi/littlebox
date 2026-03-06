// Package container provides Docker container management for agentbox.
package container

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/gbrindisi/agentbox/internal/container/builderror"
	"github.com/gbrindisi/agentbox/internal/output"
)

// Version is the current version of agentbox.
// This is used for image tagging to enable cache invalidation when agentbox is updated.
const Version = "0.1.0"

// ImageTag returns the versioned image tag for agentbox.
// Format: agentbox/base:X.X.X
func ImageTag() string {
	return fmt.Sprintf("agentbox/base:%s", Version)
}

// DerivedImageTag returns the image tag for a derived image based on the build script and workspace path.
// The tag is computed as a SHA256 hash of the build script + null byte + workspace path, using the first 12 characters.
// The null byte separator prevents collision attacks (e.g., script "A" + path "B" vs script "AB").
// Format: agentbox/build:<sha256(buildScript + "\x00" + workspacePath)[:12]>
func DerivedImageTag(buildScript string, workspacePath string) string {
	// Hash buildScript + null byte + workspacePath for workspace isolation
	hashInput := buildScript + "\x00" + workspacePath
	hash := sha256.Sum256([]byte(hashInput))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("agentbox/build:%s", hashStr[:12])
}

// EnsureImage ensures the agentbox image is available, building it if necessary.
// If forceBuild is true, the image is rebuilt even if it exists.
// The output writer receives build progress messages.
// The verbosity parameter controls output formatting: Quiet shows bullet-prefixed messages,
// Debug shows full docker build output.
func (m *Manager) EnsureImage(ctx context.Context, forceBuild bool, verbosity output.Verbosity, w io.Writer) error {
	if w == nil {
		w = io.Discard
	}

	imageTag := ImageTag()

	if !forceBuild {
		exists, err := m.ImageExists(ctx, imageTag)
		if err != nil {
			return err
		}

		if exists {
			output.BulletPrint(w, verbosity, fmt.Sprintf("Using cached image: %s", imageTag))
			return nil
		}
	}

	// In quiet mode: show "Building image... done" pattern
	// In debug mode: show full docker build output
	status := output.NewStatusWriter(w, verbosity)
	status.Start(fmt.Sprintf("Building image: %s, this may take a while", imageTag))

	buildOutput := output.Writer(w, verbosity)
	err := m.BuildImage(ctx, &BuildOptions{
		ImageName: imageTag,
		Output:    buildOutput,
	})

	if err != nil {
		status.Failed()
		return err
	}

	status.Done()
	return nil
}

// EnsureDerivedImage ensures a derived image is available, building it if necessary.
// If forceBuild is true, the image is rebuilt even if it exists.
// The output writer receives build progress messages (typically os.Stderr).
// The verbosity parameter controls output formatting: Quiet shows bullet-prefixed messages,
// Debug shows full docker build output.
// The workspacePath parameter is used for image tag computation to enable workspace isolation.
func (m *Manager) EnsureDerivedImage(ctx context.Context, buildScript string, workspacePath string, forceBuild bool, verbosity output.Verbosity, w io.Writer) (string, error) {
	if w == nil {
		w = io.Discard
	}

	derivedTag := DerivedImageTag(buildScript, workspacePath)

	if !forceBuild {
		exists, err := m.ImageExists(ctx, derivedTag)
		if err != nil {
			return "", err
		}

		if exists {
			output.BulletPrint(w, verbosity, fmt.Sprintf("Using cached derived image: %s", derivedTag))
			return derivedTag, nil
		}
	}

	// In quiet mode: show "Building derived image... done" pattern
	// In debug mode: show full docker build output
	status := output.NewStatusWriter(w, verbosity)
	status.Start(fmt.Sprintf("Building derived image: %s, this may take a while", derivedTag))

	buildOutput := output.Writer(w, verbosity)
	err := m.BuildDerivedImage(ctx, buildScript, derivedTag, buildOutput)

	if err != nil {
		status.Failed()
		return "", err
	}

	status.Done()
	return derivedTag, nil
}

// generateDerivedDockerfile generates a Dockerfile for a derived image.
// The Dockerfile extends agentbox/base and runs the build script as root.
// The image remains USER root - the entrypoint handles privilege drop to agent via setpriv.
func generateDerivedDockerfile(buildScript string) string {
	return fmt.Sprintf(`FROM %s
USER root
RUN <<'SCRIPT'
%s
SCRIPT
`, ImageTag(), buildScript)
}

// BuildDerivedImage builds a derived image from the base image with the given build script.
// The build script runs as root. The image remains USER root - the entrypoint handles
// privilege drop to agent via setpriv at runtime. Progress is written to the output writer.
func (m *Manager) BuildDerivedImage(ctx context.Context, buildScript string, imageTag string, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}

	// Generate Dockerfile
	dockerfile := generateDerivedDockerfile(buildScript)

	// Create build context (tar archive with Dockerfile)
	buildContext, err := createDerivedBuildContext(dockerfile)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	// Build the image
	// Version: BuilderBuildKit enables BuildKit features like heredoc syntax in Dockerfiles
	buildOptions := build.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
		Version:    build.BuilderBuildKit,
	}

	resp, err := m.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build derived image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Create collector for BuildKit traces
	collector := builderror.NewCollector()

	// auxCallback handles BuildKit trace messages
	auxCallback := func(msg jsonmessage.JSONMessage) {
		if msg.ID == "moby.buildkit.trace" && msg.Aux != nil {
			// Decode and collect trace
			auxBytes, err := json.Marshal(msg.Aux)
			if err == nil {
				_ = collector.Add(auxBytes)
			}
		}
	}

	// Stream and decode the build output with trace collection
	if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, output, 0, false, auxCallback); err != nil {
		// Build failed - format error with Dockerfile context
		formattedErr := collector.Format(dockerfile)
		if formattedErr != "" {
			return fmt.Errorf("build failed:\n%s", formattedErr)
		}
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

// createDerivedBuildContext creates a tar archive with the Dockerfile for a derived image.
func createDerivedBuildContext(dockerfile string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(dockerfile)
	hdr := &tar.Header{
		Name: "Dockerfile",
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tw.Write(content); err != nil {
		return nil, fmt.Errorf("failed to write tar content: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	return &buf, nil
}
