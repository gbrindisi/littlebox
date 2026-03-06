package builderror

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	controlapi "github.com/moby/buildkit/api/services/control"
	"google.golang.org/protobuf/proto"
)

// createTraceAux creates a JSONMessage.Aux field containing a base64-encoded StatusResponse.
func createTraceAux(resp *controlapi.StatusResponse) ([]byte, error) {
	// Marshal StatusResponse to protobuf bytes
	protoBytes, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}

	// Encode to base64
	base64Str := base64.StdEncoding.EncodeToString(protoBytes)

	// Wrap in JSON (Aux field is JSON-encoded base64 string)
	auxJSON, err := json.Marshal(base64Str)
	if err != nil {
		return nil, err
	}

	return auxJSON, nil
}

func TestDecodeTrace(t *testing.T) {
	// Create a sample StatusResponse
	resp := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{
				Digest: "sha256:test123",
				Name:   "RUN echo hello",
			},
		},
	}

	// Create aux JSON
	auxJSON, err := createTraceAux(resp)
	if err != nil {
		t.Fatalf("Failed to create trace aux: %v", err)
	}

	// Decode the trace
	decoded, err := DecodeTrace(auxJSON)
	if err != nil {
		t.Fatalf("DecodeTrace failed: %v", err)
	}

	// Verify decoded data
	if len(decoded.Vertexes) != 1 {
		t.Errorf("Expected 1 vertex, got %d", len(decoded.Vertexes))
	}
	if decoded.Vertexes[0].Name != "RUN echo hello" {
		t.Errorf("Expected vertex name 'RUN echo hello', got '%s'", decoded.Vertexes[0].Name)
	}
}

func TestDecodeTrace_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`not valid json`)
	_, err := DecodeTrace(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestDecodeTrace_InvalidBase64(t *testing.T) {
	// Valid JSON but invalid base64
	invalidBase64 := []byte(`"not valid base64!!!"`)
	_, err := DecodeTrace(invalidBase64)
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
}

func TestDecodeTrace_InvalidProtobuf(t *testing.T) {
	// Valid base64 but invalid protobuf
	invalidProto := base64.StdEncoding.EncodeToString([]byte("not protobuf data"))
	auxJSON, _ := json.Marshal(invalidProto)
	_, err := DecodeTrace(auxJSON)
	if err == nil {
		t.Error("Expected error for invalid protobuf, got nil")
	}
}

func TestCollector_Add(t *testing.T) {
	collector := NewCollector()

	// Create first trace
	resp1 := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{Digest: "sha256:abc", Name: "step1"},
		},
	}
	aux1, err := createTraceAux(resp1)
	if err != nil {
		t.Fatalf("Failed to create trace aux: %v", err)
	}

	// Create second trace
	resp2 := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{Digest: "sha256:def", Name: "step2"},
		},
	}
	aux2, err := createTraceAux(resp2)
	if err != nil {
		t.Fatalf("Failed to create trace aux: %v", err)
	}

	// Add traces to collector
	if err := collector.Add(aux1); err != nil {
		t.Fatalf("Failed to add first trace: %v", err)
	}
	if err := collector.Add(aux2); err != nil {
		t.Fatalf("Failed to add second trace: %v", err)
	}

	// Verify both traces were stored
	if len(collector.traces) != 2 {
		t.Errorf("Expected 2 traces, got %d", len(collector.traces))
	}
}

func TestCollector_GetError(t *testing.T) {
	tests := []struct {
		name     string
		traces   []*controlapi.StatusResponse
		expected string
	}{
		{
			name: "single error",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Name: "step1"},
						{Name: "step2", Error: "exit code 1"},
					},
				},
			},
			expected: "exit code 1",
		},
		{
			name: "multiple traces with error in second",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Name: "step1"},
					},
				},
				{
					Vertexes: []*controlapi.Vertex{
						{Name: "step2", Error: "command not found"},
					},
				},
			},
			expected: "command not found",
		},
		{
			name: "no error",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Name: "step1"},
						{Name: "step2"},
					},
				},
			},
			expected: "",
		},
		{
			name: "multiple errors returns first",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Name: "step1", Error: "first error"},
						{Name: "step2", Error: "second error"},
					},
				},
			},
			expected: "first error",
		},
		{
			name:     "empty traces",
			traces:   []*controlapi.StatusResponse{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &Collector{traces: tt.traces}
			got := collector.GetError()
			if got != tt.expected {
				t.Errorf("GetError() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCollector_GetLogs(t *testing.T) {
	tests := []struct {
		name     string
		traces   []*controlapi.StatusResponse
		expected []string
	}{
		{
			name: "single log entry",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("+ echo hello\n")},
					},
				},
			},
			expected: []string{"+ echo hello"},
		},
		{
			name: "multiple log entries",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("+ curl -fsSL example.com\n")},
						{Msg: []byte("curl: (6) Could not resolve host\n")},
					},
				},
			},
			expected: []string{
				"+ curl -fsSL example.com",
				"curl: (6) Could not resolve host",
			},
		},
		{
			name: "logs across multiple traces",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("line 1\n")},
					},
				},
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("line 2\n")},
						{Msg: []byte("line 3\n")},
					},
				},
			},
			expected: []string{"line 1", "line 2", "line 3"},
		},
		{
			name: "empty log data",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte{}},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "mixed empty and non-empty logs",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("line 1\n")},
						{Msg: []byte{}},
						{Msg: []byte("line 2\n")},
					},
				},
			},
			expected: []string{"line 1", "line 2"},
		},
		{
			name: "log without trailing newline",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("no newline")},
					},
				},
			},
			expected: []string{"no newline"},
		},
		{
			name: "only newline",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("\n")},
					},
				},
			},
			expected: []string{},
		},
		{
			name:     "empty traces",
			traces:   []*controlapi.StatusResponse{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &Collector{traces: tt.traces}
			got := collector.GetLogs()

			// Compare slices
			if len(got) != len(tt.expected) {
				t.Fatalf("GetLogs() returned %d logs, want %d\nGot: %v\nWant: %v",
					len(got), len(tt.expected), got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("GetLogs()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestCollector_Integration(t *testing.T) {
	// Simulate a build failure scenario with real-looking traces
	collector := NewCollector()

	// Trace 1: Build step starts
	resp1 := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{
				Digest: "sha256:step1",
				Name:   "RUN <<'SCRIPT'\nset -ex\ncurl -fsSL https://example.com/install.sh | bash\nSCRIPT",
			},
		},
	}
	aux1, _ := createTraceAux(resp1)
	_ = collector.Add(aux1)

	// Trace 2: Log output from the command
	resp2 := &controlapi.StatusResponse{
		Logs: []*controlapi.VertexLog{
			{Msg: []byte("+ curl -fsSL https://example.com/install.sh\n")},
			{Msg: []byte("curl: (6) Could not resolve host: example.com\n")},
		},
	}
	aux2, _ := createTraceAux(resp2)
	_ = collector.Add(aux2)

	// Trace 3: Error is reported
	resp3 := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{
				Digest: "sha256:step1",
				Error:  "process \"/bin/sh -c set -ex\\ncurl -fsSL https://example.com/install.sh | bash\\n\" did not complete successfully: exit code: 6",
			},
		},
	}
	aux3, _ := createTraceAux(resp3)
	_ = collector.Add(aux3)

	// Verify error extraction
	errorMsg := collector.GetError()
	if errorMsg == "" {
		t.Error("Expected error message, got empty string")
	}
	if errorMsg != resp3.Vertexes[0].Error {
		t.Errorf("GetError() = %q, want %q", errorMsg, resp3.Vertexes[0].Error)
	}

	// Verify log extraction
	logs := collector.GetLogs()
	expectedLogs := []string{
		"+ curl -fsSL https://example.com/install.sh",
		"curl: (6) Could not resolve host: example.com",
	}
	if len(logs) != len(expectedLogs) {
		t.Fatalf("GetLogs() returned %d logs, want %d", len(logs), len(expectedLogs))
	}
	for i := range logs {
		if logs[i] != expectedLogs[i] {
			t.Errorf("GetLogs()[%d] = %q, want %q", i, logs[i], expectedLogs[i])
		}
	}
}
