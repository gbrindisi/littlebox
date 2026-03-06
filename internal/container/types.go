package container

import (
	"io"
)

// BuildOptions specifies options for building the agentbox image.
type BuildOptions struct {
	// ImageName is the name:tag for the built image.
	// Defaults to "agentbox:latest".
	ImageName string

	// BuildArgs are additional build arguments to pass to Docker.
	BuildArgs map[string]string

	// NoCache forces a fresh build without using cache.
	NoCache bool

	// Output receives build output.
	Output io.Writer
}

// DefaultImageName is the default name for the agentbox image.
// This uses the versioned ImageTag() for cache invalidation.
var DefaultImageName = ImageTag()
