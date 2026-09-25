package container

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/gbrindisi/littlebox/internal/config"
)

// CreateContainer creates a new container with the configuration from an Agentfile.
// The container is created with workspace and additional mounts, environment variables,
// and network settings from the config. The tty parameter controls whether to allocate
// a TTY for the container (use DetectTTY to determine this from a TTYMode).
// The imageTag parameter specifies which image to use for the container.
// It returns the container ID.
func CreateContainer(ctx context.Context, cli *client.Client, cfg *config.Config, args []string, tty bool, imageTag string) (string, error) {
	return CreateContainerWithOptions(ctx, cli, cfg, args, tty, imageTag, CreateOptions{})
}

// CreateContainerWithOptions is like CreateContainer but also applies an optional
// container name and labels. The label littlebox=1 is always set.
func CreateContainerWithOptions(ctx context.Context, cli *client.Client, cfg *config.Config, args []string, tty bool, imageTag string, opts CreateOptions) (string, error) {
	if err := ValidateContainerName(opts.Name); err != nil {
		return "", err
	}
	mounts := BuildMounts(cfg)

	env := config.ResolveEnvPassthrough(cfg.Agent.EnvPassthrough)
	env = append(env, "ALLOWED_DOMAINS="+strings.Join(cfg.Network.Allow, ","))

	cmd := append([]string{}, cfg.Agent.Command...)
	cmd = append(cmd, cfg.Agent.Args...)
	cmd = append(cmd, args...)

	// Build security options
	var securityOpt []string
	if cfg.Container.NoNewPrivileges != nil && *cfg.Container.NoNewPrivileges {
		securityOpt = append(securityOpt, "no-new-privileges")
	}

	// Add seccomp profile if configured
	// Default ("" or "default") = Docker's default seccomp profile (nothing to add)
	// "unconfined" = disable seccomp filtering
	// custom path = use custom seccomp profile
	seccomp := cfg.Container.SeccompProfile
	if seccomp != "" && seccomp != "default" {
		securityOpt = append(securityOpt, "seccomp="+seccomp)
	}

	// Build host config
	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		AutoRemove:  true,
		SecurityOpt: securityOpt,
		CapAdd:      []string{"NET_ADMIN"},
	}

	// Read-only root filesystem with tmpfs for writable directories
	if cfg.Container.ReadOnlyRoot {
		hostConfig.ReadonlyRootfs = true
		hostConfig.Tmpfs = map[string]string{
			"/tmp":        "rw,noexec,nosuid",
			"/home/agent": "rw,noexec,nosuid",
		}
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:        imageTag,
			Cmd:          cmd,
			Env:          env,
			WorkingDir:   "/workspace",
			Tty:          tty,
			Labels:       containerLabels(opts.Labels),
			OpenStdin:    true,
			StdinOnce:    true, // Close stdin after first client disconnects
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		},
		hostConfig,
		nil, nil, opts.Name,
	)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}
