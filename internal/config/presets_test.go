package config

import (
	"slices"
	"testing"
)

func TestNetworkPresetsExist(t *testing.T) {
	// Service-specific presets should exist
	presets := []string{
		"anthropic", "openai", "google-ai", "mistral",
		"github", "gitlab", "bitbucket",
		"npm", "pypi", "cargo", "rubygems",
		"huggingface",
	}
	for _, preset := range presets {
		if _, ok := NetworkPresets[preset]; !ok {
			t.Errorf("expected preset %q to exist in NetworkPresets", preset)
		}
	}
}

func TestAnthropicPreset(t *testing.T) {
	domains := NetworkPresets["anthropic"]
	expected := []string{"api.anthropic.com", "anthropic.com", "claude.ai"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("anthropic preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestGitHubPreset(t *testing.T) {
	domains := NetworkPresets["github"]
	expected := []string{
		"github.com",
		"api.github.com",
		"raw.githubusercontent.com",
		"objects.githubusercontent.com",
		"codeload.github.com",
		"gist.githubusercontent.com",
	}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("github preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestNpmPreset(t *testing.T) {
	domains := NetworkPresets["npm"]
	expected := []string{"registry.npmjs.org", "npmjs.org", "npmjs.com"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("npm preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestPypiPreset(t *testing.T) {
	domains := NetworkPresets["pypi"]
	expected := []string{"pypi.org", "files.pythonhosted.org"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("pypi preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestOpenAIPreset(t *testing.T) {
	domains := NetworkPresets["openai"]
	expected := []string{"api.openai.com", "openai.com", "platform.openai.com", "cdn.openai.com"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("openai preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestGoogleAIPreset(t *testing.T) {
	domains := NetworkPresets["google-ai"]
	expected := []string{"generativelanguage.googleapis.com", "ai.google.dev", "aistudio.google.com"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("google-ai preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestMistralPreset(t *testing.T) {
	domains := NetworkPresets["mistral"]
	expected := []string{"api.mistral.ai", "mistral.ai"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("mistral preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestGitLabPreset(t *testing.T) {
	domains := NetworkPresets["gitlab"]
	expected := []string{"gitlab.com", "registry.gitlab.com"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("gitlab preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestBitbucketPreset(t *testing.T) {
	domains := NetworkPresets["bitbucket"]
	expected := []string{"bitbucket.org", "api.bitbucket.org"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("bitbucket preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestCargoPreset(t *testing.T) {
	domains := NetworkPresets["cargo"]
	expected := []string{"crates.io", "static.crates.io", "index.crates.io"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("cargo preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestRubygemsPreset(t *testing.T) {
	domains := NetworkPresets["rubygems"]
	expected := []string{"rubygems.org"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("rubygems preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestHuggingfacePreset(t *testing.T) {
	domains := NetworkPresets["huggingface"]
	expected := []string{"huggingface.co", "cdn-lfs.huggingface.co"}

	for _, domain := range expected {
		if !slices.Contains(domains, domain) {
			t.Errorf("huggingface preset: expected domain %q to be allowed", domain)
		}
	}
}

func TestApplyNetworkPresetsSingle(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{
			Presets: []string{"anthropic"},
		},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	if !slices.Contains(cfg.Network.Allow, "anthropic.com") {
		t.Error("anthropic preset should allow anthropic.com")
	}
	if !slices.Contains(cfg.Network.Allow, "api.anthropic.com") {
		t.Error("anthropic preset should allow api.anthropic.com")
	}
	if slices.Contains(cfg.Network.Allow, "github.com") {
		t.Error("anthropic preset should not allow github.com")
	}
}

func TestApplyNetworkPresetsMultiple(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{
			Presets: []string{"anthropic", "github", "npm"},
		},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	expectedDomains := []string{
		"anthropic.com",
		"api.anthropic.com",
		"github.com",
		"api.github.com",
		"registry.npmjs.org",
	}

	for _, domain := range expectedDomains {
		if !slices.Contains(cfg.Network.Allow, domain) {
			t.Errorf("combined presets should allow %s", domain)
		}
	}
}

func TestApplyNetworkPresetsAppendsAdditionalAllows(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{
			Presets: []string{"anthropic"},
			Allow:   []string{"custom.example.com"},
		},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	// Should have preset domains
	if !slices.Contains(cfg.Network.Allow, "anthropic.com") {
		t.Error("should include preset domain anthropic.com")
	}

	// Should have additional allow entries
	if !slices.Contains(cfg.Network.Allow, "custom.example.com") {
		t.Error("should include additional allow entry custom.example.com")
	}
}

func TestApplyNetworkPresetsUnknownPreset(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{
			Presets: []string{"unknown"},
		},
	}

	err := ApplyNetworkPresets(cfg)
	if err == nil {
		t.Error("expected error for unknown preset")
	}
	if err.Error() != "unknown network preset: unknown" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestApplyNetworkPresetsNoPreset(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{
			Allow: []string{"example.com"},
		},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	// Should not modify the allow list
	if len(cfg.Network.Allow) != 1 {
		t.Errorf("expected 1 allowed domain, got %d", len(cfg.Network.Allow))
	}
	if cfg.Network.Allow[0] != "example.com" {
		t.Errorf("expected example.com, got %s", cfg.Network.Allow[0])
	}
}

func TestApplyNetworkPresetsEmptyConfig(t *testing.T) {
	cfg := &Config{
		Network: NetworkConfig{},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	// Should have empty allow list (air-gapped execution is valid)
	if len(cfg.Network.Allow) != 0 {
		t.Errorf("expected 0 allowed domains, got %d", len(cfg.Network.Allow))
	}
}

func TestMixedAllowListEntries(t *testing.T) {
	// Test that allow list can contain a mix of domains, IPs, and CIDRs
	// Each entry type should be correctly identified
	cfg := &Config{
		Network: NetworkConfig{
			Presets: []string{"anthropic"},
			Allow: []string{
				"custom.example.com", // domain
				"10.0.0.1",           // IP
				"192.168.1.0/24",     // CIDR
				"api.internal.local", // domain
				"172.16.0.100",       // IP
				"10.10.0.0/16",       // CIDR
			},
		},
	}

	if err := ApplyNetworkPresets(cfg); err != nil {
		t.Fatalf("ApplyNetworkPresets failed: %v", err)
	}

	// Verify preset domains are included
	if !slices.Contains(cfg.Network.Allow, "anthropic.com") {
		t.Error("should include preset domain anthropic.com")
	}

	// Verify all mixed allow entries are preserved
	expectedEntries := []string{
		"custom.example.com",
		"10.0.0.1",
		"192.168.1.0/24",
		"api.internal.local",
		"172.16.0.100",
		"10.10.0.0/16",
	}
	for _, entry := range expectedEntries {
		if !slices.Contains(cfg.Network.Allow, entry) {
			t.Errorf("should include allow entry %s", entry)
		}
	}

	// Verify each entry type is correctly detected
	for _, entry := range expectedEntries {
		entryType := DetectEntryType(entry)
		switch entry {
		case "custom.example.com", "api.internal.local":
			if entryType != EntryTypeDomain {
				t.Errorf("expected %s to be detected as domain, got %s", entry, entryType)
			}
		case "10.0.0.1", "172.16.0.100":
			if entryType != EntryTypeIP {
				t.Errorf("expected %s to be detected as IP, got %s", entry, entryType)
			}
		case "192.168.1.0/24", "10.10.0.0/16":
			if entryType != EntryTypeCIDR {
				t.Errorf("expected %s to be detected as CIDR, got %s", entry, entryType)
			}
		}
	}
}
