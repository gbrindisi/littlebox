package container

import (
	"path/filepath"

	"github.com/docker/docker/api/types/mount"

	"github.com/gbrindisi/agentbox/internal/config"
)

// BuildMounts creates the mount configuration for a container.
// It builds workspace, additional, and redacted file mounts from the config.
func BuildMounts(cfg *config.Config) []mount.Mount {
	var mounts []mount.Mount

	// Workspace mount
	if cfg.Workspace.Path != "" {
		readOnly := false
		if cfg.Workspace.Writable != nil {
			readOnly = !*cfg.Workspace.Writable
		}
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   cfg.Workspace.Path,
			Target:   cfg.Workspace.MountPoint,
			ReadOnly: readOnly,
		})
	}

	// Additional mounts from config
	for _, m := range cfg.Mounts {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// Redacted file mounts (tmpfs overlays to hide sensitive files)
	// Use GetRedactedFiles to properly resolve glob patterns
	if cfg.Workspace.Path != "" && len(cfg.Redact) > 0 {
		redactedFiles, _ := config.GetRedactedFiles(cfg.Workspace.Path, cfg.Redact)
		for _, absPath := range redactedFiles {
			relPath, err := filepath.Rel(cfg.Workspace.Path, absPath)
			if err != nil {
				continue
			}
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeTmpfs,
				Target: filepath.Join(cfg.Workspace.MountPoint, relPath),
			})
		}
	}

	return mounts
}
