package container

import (
	"testing"

	controlapi "github.com/moby/buildkit/api/services/control"
	"google.golang.org/protobuf/proto"
)

// TestBuildKitDependency verifies that we can import and use BuildKit controlapi types.
// This validates that the moby/buildkit dependency is correctly installed and that
// protobuf decoding works with the current Docker API version.
func TestBuildKitDependency(t *testing.T) {
	// Create a sample StatusResponse to verify the import works
	resp := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{
				Digest: "sha256:abc123",
				Name:   "test-vertex",
			},
		},
	}

	// Marshal to protobuf bytes
	data, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal StatusResponse: %v", err)
	}

	// Unmarshal back to verify round-trip works
	decoded := &controlapi.StatusResponse{}
	err = proto.Unmarshal(data, decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal StatusResponse: %v", err)
	}

	// Verify the data
	if len(decoded.Vertexes) != 1 {
		t.Errorf("Expected 1 vertex, got %d", len(decoded.Vertexes))
	}
	if decoded.Vertexes[0].Name != "test-vertex" {
		t.Errorf("Expected vertex name 'test-vertex', got '%s'", decoded.Vertexes[0].Name)
	}
}
