package config

import "fmt"

// NetworkPresets maps preset names to their allowed domain lists.
// Each preset represents a logical service (AI provider, code host, package registry).
// Presets are composable - users can combine multiple presets in their config.
var NetworkPresets = map[string][]string{
	// AI providers
	"anthropic": {
		"api.anthropic.com",
		"anthropic.com",
		"claude.ai",
	},
	"openai": {
		"api.openai.com",
		"openai.com",
		"platform.openai.com",
		"cdn.openai.com",
	},
	"google-ai": {
		"generativelanguage.googleapis.com",
		"ai.google.dev",
		"aistudio.google.com",
	},
	"mistral": {
		"api.mistral.ai",
		"mistral.ai",
	},

	// Source control
	"github": {
		"github.com",
		"api.github.com",
		"raw.githubusercontent.com",
		"objects.githubusercontent.com",
		"codeload.github.com",
		"gist.githubusercontent.com",
	},
	"gitlab": {
		"gitlab.com",
		"registry.gitlab.com",
	},
	"bitbucket": {
		"bitbucket.org",
		"api.bitbucket.org",
	},

	// Package registries
	"npm": {
		"registry.npmjs.org",
		"npmjs.org",
		"npmjs.com",
	},
	"pypi": {
		"pypi.org",
		"files.pythonhosted.org",
	},
	"cargo": {
		"crates.io",
		"static.crates.io",
		"index.crates.io",
	},
	"rubygems": {
		"rubygems.org",
	},

	// ML/AI models
	"huggingface": {
		"huggingface.co",
		"cdn-lfs.huggingface.co",
	},
}

// ApplyNetworkPresets applies network presets to the config.
// It iterates over all presets in cfg.Network.Presets and combines their domains.
// Any additional allow entries in the config are appended to the preset domains.
func ApplyNetworkPresets(cfg *Config) error {
	if len(cfg.Network.Presets) == 0 {
		return nil
	}

	var domains []string
	for _, presetName := range cfg.Network.Presets {
		presetDomains, ok := NetworkPresets[presetName]
		if !ok {
			return fmt.Errorf("unknown network preset: %s", presetName)
		}
		domains = append(domains, presetDomains...)
	}

	// Preserve any additional allow entries from the config
	additionalAllows := cfg.Network.Allow
	cfg.Network.Allow = append(domains, additionalAllows...)

	return nil
}
