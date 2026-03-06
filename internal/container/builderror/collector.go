package builderror

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	controlapi "github.com/moby/buildkit/api/services/control"
	"google.golang.org/protobuf/proto"
)

// Collector accumulates BuildKit trace messages during a Docker build.
// It decodes StatusResponse protobuf messages and stores them for later
// error formatting if the build fails.
type Collector struct {
	traces []*controlapi.StatusResponse
}

// NewCollector creates a new trace collector.
func NewCollector() *Collector {
	return &Collector{
		traces: make([]*controlapi.StatusResponse, 0),
	}
}

// Add decodes a BuildKit trace from JSONMessage.Aux and stores it.
// The aux parameter should be the raw JSON bytes from JSONMessage.Aux field
// when JSONMessage.ID == "moby.buildkit.trace".
func (c *Collector) Add(aux []byte) error {
	trace, err := DecodeTrace(aux)
	if err != nil {
		return err
	}
	c.traces = append(c.traces, trace)
	return nil
}

// DecodeTrace decodes a BuildKit trace from base64-encoded JSON to StatusResponse.
// The input is the raw Aux field bytes which contain a base64-encoded protobuf message.
func DecodeTrace(auxJSON []byte) (*controlapi.StatusResponse, error) {
	// The Aux field contains base64-encoded protobuf bytes
	var base64Str string
	if err := json.Unmarshal(auxJSON, &base64Str); err != nil {
		return nil, fmt.Errorf("failed to unmarshal aux JSON: %w", err)
	}

	// Decode base64 to protobuf bytes
	protoBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Unmarshal protobuf to StatusResponse
	trace := &controlapi.StatusResponse{}
	if err := proto.Unmarshal(protoBytes, trace); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	return trace, nil
}

// GetError returns the first non-empty error message from collected traces.
// It iterates through all Vertex entries looking for the Error field.
// Returns empty string if no error is found.
func (c *Collector) GetError() string {
	for _, trace := range c.traces {
		for _, vertex := range trace.Vertexes {
			if vertex.Error != "" {
				return vertex.Error
			}
		}
	}
	return ""
}

// GetLogs returns all build output lines from collected traces.
// It extracts VertexLog.Msg entries which contain the command output
// (stdout/stderr) from the build process.
func (c *Collector) GetLogs() []string {
	var logs []string
	for _, trace := range c.traces {
		for _, log := range trace.Logs {
			if len(log.Msg) > 0 {
				// VertexLog.Msg is raw bytes, convert to string
				// Trim trailing newlines as each log entry typically has one
				line := strings.TrimRight(string(log.Msg), "\n")
				if line != "" {
					logs = append(logs, line)
				}
			}
		}
	}
	return logs
}
