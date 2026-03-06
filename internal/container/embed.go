// Package container provides Docker container management for agentbox.
package container

import (
	"embed"
	"io/fs"
)

//go:embed docker/*
var dockerAssets embed.FS

// DockerAssets returns the embedded Docker assets filesystem.
// This includes the Dockerfile, entrypoint script, and firewall script.
func DockerAssets() fs.FS {
	sub, err := fs.Sub(dockerAssets, "docker")
	if err != nil {
		// This should never happen with a valid embed directive
		panic("failed to access embedded docker assets: " + err.Error())
	}
	return sub
}

// GetDockerfile returns the embedded Dockerfile contents.
func GetDockerfile() ([]byte, error) {
	return dockerAssets.ReadFile("docker/Dockerfile")
}

// GetEntrypoint returns the embedded entrypoint.sh contents.
func GetEntrypoint() ([]byte, error) {
	return dockerAssets.ReadFile("docker/entrypoint.sh")
}

// GetInitFirewall returns the embedded init-firewall.sh contents.
// This is the new firewall script with ipset, DNS preservation, and verification.
func GetInitFirewall() ([]byte, error) {
	return dockerAssets.ReadFile("docker/init-firewall.sh")
}

// GetLibsandbox returns the embedded libsandbox.c contents.
// This is the LD_PRELOAD library for friendly firewall error messages.
func GetLibsandbox() ([]byte, error) {
	return dockerAssets.ReadFile("docker/libsandbox.c")
}
