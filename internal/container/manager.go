package container

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/gbrindisi/littlebox/internal/container/builderror"
)

// Manager handles Docker container lifecycle operations.
type Manager struct {
	client *client.Client
}

// NewManager creates a new container manager.
// It connects to the Docker daemon using environment variables or default socket.
func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Manager{client: cli}, nil
}

// NewManagerWithClient creates a new container manager with a specific Docker client.
// This is useful for testing or when custom client configuration is needed.
func NewManagerWithClient(cli *client.Client) *Manager {
	return &Manager{client: cli}
}

// Close releases resources held by the manager.
func (m *Manager) Close() error {
	return m.client.Close()
}

// Client returns the underlying Docker client.
// This is useful for using standalone functions that require a *client.Client.
func (m *Manager) Client() *client.Client {
	return m.client
}

// BuildImage builds the littlebox Docker image from embedded assets.
func (m *Manager) BuildImage(ctx context.Context, opts *BuildOptions) error {
	if opts == nil {
		opts = &BuildOptions{}
	}

	imageName := opts.ImageName
	if imageName == "" {
		imageName = DefaultImageName
	}

	buildContext, err := createBuildContext()
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	buildArgs := make(map[string]*string)
	for k, v := range opts.BuildArgs {
		value := v
		buildArgs[k] = &value
	}

	buildOptions := build.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: "Dockerfile",
		BuildArgs:  buildArgs,
		NoCache:    opts.NoCache,
		Remove:     true,
	}

	resp, err := m.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	output := opts.Output
	if output == nil {
		output = io.Discard
	}

	// Get Dockerfile content for error formatting
	dockerfileBytes, err := GetDockerfile()
	if err != nil {
		return fmt.Errorf("failed to get dockerfile: %w", err)
	}
	dockerfile := string(dockerfileBytes)

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

// ImageExists checks if an image exists locally.
func (m *Manager) ImageExists(ctx context.Context, imageName string) (bool, error) {
	_, err := m.client.ImageInspect(ctx, imageName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect image %s: %w", imageName, err)
	}
	return true, nil
}

// createBuildContext creates a tar archive with all Docker build files.
func createBuildContext() (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := map[string]func() ([]byte, error){
		"Dockerfile":       GetDockerfile,
		"entrypoint.sh":    GetEntrypoint,
		"init-firewall.sh": GetInitFirewall,
		"libsandbox.c":     GetLibsandbox,
	}

	for name, getContent := range files {
		content, err := getContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", name, err)
		}

		mode := int64(0644)
		if strings.HasSuffix(name, ".sh") {
			mode = 0755
		}

		hdr := &tar.Header{
			Name: name,
			Mode: mode,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("failed to write tar header for %s: %w", name, err)
		}

		if _, err := tw.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write tar content for %s: %w", name, err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}

	return &buf, nil
}
