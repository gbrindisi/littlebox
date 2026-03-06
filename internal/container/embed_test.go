package container

import (
	"io/fs"
	"strings"
	"testing"
)

func TestDockerAssets(t *testing.T) {
	assets := DockerAssets()

	// Check that we can list files
	entries, err := fs.ReadDir(assets, ".")
	if err != nil {
		t.Fatalf("failed to read docker assets directory: %v", err)
	}

	// Verify expected files are present
	expectedFiles := map[string]bool{
		"Dockerfile":       false,
		"entrypoint.sh":    false,
		"init-firewall.sh": false,
	}

	for _, entry := range entries {
		if _, ok := expectedFiles[entry.Name()]; ok {
			expectedFiles[entry.Name()] = true
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("expected file %q not found in docker assets", name)
		}
	}
}

func TestGetDockerfile(t *testing.T) {
	content, err := GetDockerfile()
	if err != nil {
		t.Fatalf("GetDockerfile() error = %v", err)
	}

	// Verify Dockerfile content
	s := string(content)
	if !strings.Contains(s, "FROM debian:bookworm-slim") {
		t.Error("Dockerfile missing base image")
	}
	if !strings.Contains(s, "iptables") {
		t.Error("Dockerfile missing iptables package")
	}
	if !strings.Contains(s, "ipset") {
		t.Error("Dockerfile missing ipset package")
	}
	if !strings.Contains(s, "dnsutils") {
		t.Error("Dockerfile missing dnsutils package")
	}
	if !strings.Contains(s, "jq") {
		t.Error("Dockerfile missing jq package")
	}
	if !strings.Contains(s, "useradd") {
		t.Error("Dockerfile missing user creation")
	}
	if !strings.Contains(s, "ENTRYPOINT") {
		t.Error("Dockerfile missing entrypoint")
	}
}

func TestGetEntrypoint(t *testing.T) {
	content, err := GetEntrypoint()
	if err != nil {
		t.Fatalf("GetEntrypoint() error = %v", err)
	}

	s := string(content)
	if !strings.HasPrefix(s, "#!/bin/bash") {
		t.Error("entrypoint.sh missing shebang")
	}
	if !strings.Contains(s, "init-firewall.sh") {
		t.Error("entrypoint.sh missing init-firewall.sh call")
	}
	if !strings.Contains(s, "ALLOWED_DOMAINS") {
		t.Error("entrypoint.sh missing ALLOWED_DOMAINS check")
	}
	if !strings.Contains(s, "exec") {
		t.Error("entrypoint.sh missing exec")
	}
	if !strings.Contains(s, "setpriv") {
		t.Error("entrypoint.sh missing setpriv privilege drop")
	}
	if !strings.Contains(s, "--reuid") {
		t.Error("entrypoint.sh missing --reuid for user transition")
	}
	if !strings.Contains(s, "--regid") {
		t.Error("entrypoint.sh missing --regid for group transition")
	}
}

func TestGetInitFirewall(t *testing.T) {
	content, err := GetInitFirewall()
	if err != nil {
		t.Fatalf("GetInitFirewall() error = %v", err)
	}

	s := string(content)
	if !strings.HasPrefix(s, "#!/bin/bash") {
		t.Error("init-firewall.sh missing shebang")
	}
	if !strings.Contains(s, "iptables") {
		t.Error("init-firewall.sh missing iptables commands")
	}
	if !strings.Contains(s, "ipset") {
		t.Error("init-firewall.sh missing ipset commands")
	}
	if !strings.Contains(s, "ALLOWED_DOMAINS") {
		t.Error("init-firewall.sh missing ALLOWED_DOMAINS environment variable")
	}
	if !strings.Contains(s, "DNS_SERVER") {
		t.Error("init-firewall.sh missing DNS_SERVER handling")
	}
	if !strings.Contains(s, "dig") {
		t.Error("init-firewall.sh missing dig for domain resolution")
	}
	if !strings.Contains(s, "github.com/meta") {
		t.Error("init-firewall.sh missing GitHub IP fetching")
	}
	if !strings.Contains(s, "verify_firewall") {
		t.Error("init-firewall.sh missing firewall verification")
	}
}
