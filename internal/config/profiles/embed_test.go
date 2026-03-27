package profiles

import (
	"strings"
	"testing"

	"github.com/gbrindisi/littlebox/internal/config"
	"gopkg.in/yaml.v3"
)

func TestNames(t *testing.T) {
	names := Names()

	// Verify we got some names
	if len(names) == 0 {
		t.Fatal("Names() returned empty list")
	}

	// Verify expected profiles are present
	expectedProfiles := []string{"claude-code", "codex-cli", "openhands"}
	foundProfiles := make(map[string]bool)
	for _, name := range names {
		foundProfiles[name] = true
	}

	for _, expected := range expectedProfiles {
		if !foundProfiles[expected] {
			t.Errorf("expected profile %q not found in names list", expected)
		}
	}

	// Verify names are sorted
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Errorf("names not sorted: %q >= %q at index %d", names[i-1], names[i], i)
		}
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name       string
		profile    string
		wantFound  bool
		wantPrefix string // Expected content prefix for validation
	}{
		{
			name:       "claude-code profile exists",
			profile:    "claude-code",
			wantFound:  true,
			wantPrefix: "# agentbox configuration",
		},
		{
			name:       "codex-cli profile exists",
			profile:    "codex-cli",
			wantFound:  true,
			wantPrefix: "# agentbox configuration",
		},
		{
			name:       "openhands profile exists",
			profile:    "openhands",
			wantFound:  true,
			wantPrefix: "# agentbox configuration",
		},
		{
			name:      "non-existent profile",
			profile:   "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, found := Get(tt.profile)

			if found != tt.wantFound {
				t.Errorf("Get(%q) found = %v, want %v", tt.profile, found, tt.wantFound)
			}

			if tt.wantFound {
				if content == "" {
					t.Errorf("Get(%q) returned empty content", tt.profile)
				}
				if tt.wantPrefix != "" && !strings.HasPrefix(content, tt.wantPrefix) {
					t.Errorf("Get(%q) content doesn't start with expected prefix", tt.profile)
				}
			} else {
				if content != "" {
					t.Errorf("Get(%q) should return empty string when not found", tt.profile)
				}
			}
		})
	}
}

func TestProfileContent(t *testing.T) {
	// Verify that all profiles returned by Names() can be retrieved
	names := Names()
	for _, name := range names {
		content, found := Get(name)
		if !found {
			t.Errorf("profile %q in Names() but Get returned not found", name)
		}
		if content == "" {
			t.Errorf("profile %q has empty content", name)
		}
	}
}

// TestAllProfiles validates all embedded profiles dynamically:
// - Iterates over all discovered profiles using Names()
// - Ensures each profile can be retrieved and has non-empty content
// - Parses content as YAML into config.Config struct
// - Verifies YAML parsing succeeds (valid config structure)
func TestAllProfiles(t *testing.T) {
	names := Names()

	// Ensure we have some profiles
	if len(names) == 0 {
		t.Fatal("no profiles discovered")
	}

	// Validate each profile
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			// 1. Profile must exist
			content, found := Get(name)
			if !found {
				t.Fatalf("profile %q not found", name)
			}

			// 2. Content must be non-empty
			if content == "" {
				t.Fatalf("profile %q has empty content", name)
			}

			// 3. Content must parse as valid YAML config
			var cfg config.Config
			err := yaml.Unmarshal([]byte(content), &cfg)
			if err != nil {
				t.Fatalf("profile %q failed to parse as YAML: %v", name, err)
			}

			// 4. Basic sanity check: agent section should be present
			// (all profiles should define an agent)
			if len(cfg.Agent.Command) == 0 {
				t.Errorf("profile %q has empty agent command", name)
			}
		})
	}

	// 5. Verify expected profiles are present
	expectedProfiles := map[string]bool{
		"claude-code": false,
		"codex-cli":   false,
		"openhands":   false,
	}

	for _, name := range names {
		if _, ok := expectedProfiles[name]; ok {
			expectedProfiles[name] = true
		}
	}

	for profile, found := range expectedProfiles {
		if !found {
			t.Errorf("expected profile %q not found in discovered profiles", profile)
		}
	}
}
