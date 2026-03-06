package container

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

func TestCreateBuildContext(t *testing.T) {
	reader, err := createBuildContext()
	if err != nil {
		t.Fatalf("createBuildContext() error = %v", err)
	}

	// Read the tar archive
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("failed to read tar: %v", err)
	}

	// Parse the tar and verify contents
	tr := tar.NewReader(&buf)
	files := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar header: %v", err)
		}
		files[hdr.Name] = true

		// Verify shell scripts are executable
		if hdr.Name == "entrypoint.sh" || hdr.Name == "init-firewall.sh" {
			if hdr.Mode != 0755 {
				t.Errorf("expected %s mode 0755, got %o", hdr.Name, hdr.Mode)
			}
		}

		// Verify non-shell files have regular mode
		if hdr.Name == "Dockerfile" {
			if hdr.Mode != 0644 {
				t.Errorf("expected %s mode 0644, got %o", hdr.Name, hdr.Mode)
			}
		}
	}

	// Verify all expected files are present
	expected := []string{"Dockerfile", "entrypoint.sh", "init-firewall.sh"}
	for _, name := range expected {
		if !files[name] {
			t.Errorf("expected file %s not found in build context", name)
		}
	}
}

func TestDefaultConstants(t *testing.T) {
	// DefaultImageName now uses the versioned ImageTag()
	if DefaultImageName != ImageTag() {
		t.Errorf("expected DefaultImageName to be %s, got %s", ImageTag(), DefaultImageName)
	}
}
