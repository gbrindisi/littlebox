package builderror

import (
	"strings"
	"testing"

	controlapi "github.com/moby/buildkit/api/services/control"
)

func TestFormatDockerfile(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		failedLine int
		want       string
	}{
		{
			name:       "simple dockerfile no failed line",
			content:    "FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl",
			failedLine: 0,
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN apt-get update\n" +
				" 3 | RUN apt-get install -y curl\n",
		},
		{
			name:       "highlight line 2",
			content:    "FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl",
			failedLine: 2,
			want: " 1 | FROM ubuntu:22.04\n" +
				"> 2 | RUN apt-get update\n" +
				" 3 | RUN apt-get install -y curl\n",
		},
		{
			name:       "highlight line 3",
			content:    "FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl",
			failedLine: 3,
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN apt-get update\n" +
				"> 3 | RUN apt-get install -y curl\n",
		},
		{
			name:       "single line dockerfile",
			content:    "FROM ubuntu:22.04",
			failedLine: 1,
			want:       "> 1 | FROM ubuntu:22.04\n",
		},
		{
			name:       "empty content",
			content:    "",
			failedLine: 0,
			want:       "",
		},
		{
			name:       "multiline RUN command",
			content:    "FROM ubuntu:22.04\nRUN <<'SCRIPT'\nset -ex\ncurl https://example.com\nSCRIPT",
			failedLine: 2,
			want: " 1 | FROM ubuntu:22.04\n" +
				"> 2 | RUN <<'SCRIPT'\n" +
				" 3 | set -ex\n" +
				" 4 | curl https://example.com\n" +
				" 5 | SCRIPT\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDockerfile(tt.content, tt.failedLine)
			if got != tt.want {
				t.Errorf("formatDockerfile() mismatch\nGot:\n%s\nWant:\n%s", got, tt.want)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name string
		logs []string
		want string
	}{
		{
			name: "empty logs",
			logs: []string{},
			want: "",
		},
		{
			name: "single line",
			logs: []string{"+ echo hello"},
			want: "  + echo hello\n",
		},
		{
			name: "exactly 20 lines - no truncation",
			logs: func() []string {
				lines := make([]string, 20)
				for i := 0; i < 20; i++ {
					lines[i] = "line"
				}
				return lines
			}(),
			want: strings.Repeat("  line\n", 20),
		},
		{
			name: "21 lines - show last 20 with truncation hint",
			logs: func() []string {
				lines := make([]string, 21)
				for i := 0; i < 21; i++ {
					lines[i] = "line"
				}
				return lines
			}(),
			want: strings.Repeat("  line\n", 20) + "\n... (1 more lines, use --debug for full output)\n",
		},
		{
			name: "50 lines - show last 20 with truncation hint",
			logs: func() []string {
				lines := make([]string, 50)
				for i := 0; i < 50; i++ {
					lines[i] = "line"
				}
				return lines
			}(),
			want: strings.Repeat("  line\n", 20) + "\n... (30 more lines, use --debug for full output)\n",
		},
		{
			name: "multiple different lines with truncation",
			logs: func() []string {
				lines := make([]string, 25)
				for i := 0; i < 25; i++ {
					lines[i] = string(rune('a' + i))
				}
				return lines
			}(),
			want: func() string {
				var result strings.Builder
				// Last 20 lines start from index 5 (f)
				for i := 5; i < 25; i++ {
					result.WriteString("  " + string(rune('a'+i)) + "\n")
				}
				result.WriteString("\n... (5 more lines, use --debug for full output)\n")
				return result.String()
			}(),
		},
		{
			name: "real curl error output",
			logs: []string{
				"+ curl -fsSL https://example.com/install.sh",
				"curl: (6) Could not resolve host: example.com",
			},
			want: "  + curl -fsSL https://example.com/install.sh\n" +
				"  curl: (6) Could not resolve host: example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOutput(tt.logs)
			if got != tt.want {
				t.Errorf("formatOutput() mismatch\nGot:\n%q\nWant:\n%q", got, tt.want)
			}
		})
	}
}

func TestCollector_Format(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		traces     []*controlapi.StatusResponse
		want       string
	}{
		{
			name: "simple error with short output",
			dockerfile: "FROM ubuntu:22.04\n" +
				"RUN apt-get update",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Error: "exit code 1"},
					},
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("+ apt-get update\n")},
						{Msg: []byte("E: Unable to locate package\n")},
					},
				},
			},
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN apt-get update\n" +
				"\n" +
				"Error: exit code 1\n" +
				"\n" +
				"  + apt-get update\n" +
				"  E: Unable to locate package\n",
		},
		{
			name: "error with no logs",
			dockerfile: "FROM ubuntu:22.04\n" +
				"RUN exit 1",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Error: "exit code 1"},
					},
				},
			},
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN exit 1\n" +
				"\n" +
				"Error: exit code 1\n" +
				"\n",
		},
		{
			name: "no error but has logs",
			dockerfile: "FROM ubuntu:22.04\n" +
				"RUN echo hello",
			traces: []*controlapi.StatusResponse{
				{
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("+ echo hello\n")},
						{Msg: []byte("hello\n")},
					},
				},
			},
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN echo hello\n" +
				"\n" +
				"  + echo hello\n" +
				"  hello\n",
		},
		{
			name: "long output gets truncated",
			dockerfile: "FROM ubuntu:22.04\n" +
				"RUN generate_output",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Error: "exit code 1"},
					},
					Logs: func() []*controlapi.VertexLog {
						logs := make([]*controlapi.VertexLog, 30)
						for i := 0; i < 30; i++ {
							logs[i] = &controlapi.VertexLog{
								Msg: []byte("line\n"),
							}
						}
						return logs
					}(),
				},
			},
			want: " 1 | FROM ubuntu:22.04\n" +
				" 2 | RUN generate_output\n" +
				"\n" +
				"Error: exit code 1\n" +
				"\n" +
				strings.Repeat("  line\n", 20) +
				"\n... (10 more lines, use --debug for full output)\n",
		},
		{
			name:       "empty dockerfile",
			dockerfile: "",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Error: "build failed"},
					},
				},
			},
			want: "\n" +
				"Error: build failed\n" +
				"\n",
		},
		{
			name: "multiline script error",
			dockerfile: "FROM agentbox/base:0.1.0\n" +
				"USER root\n" +
				"RUN <<'SCRIPT'\n" +
				"set -ex\n" +
				"curl -fsSL https://example.com/install.sh | bash\n" +
				"SCRIPT",
			traces: []*controlapi.StatusResponse{
				{
					Vertexes: []*controlapi.Vertex{
						{Error: "exit code 1"},
					},
					Logs: []*controlapi.VertexLog{
						{Msg: []byte("+ curl -fsSL https://example.com/install.sh | bash\n")},
						{Msg: []byte("curl: (6) Could not resolve host: example.com\n")},
					},
				},
			},
			want: " 1 | FROM agentbox/base:0.1.0\n" +
				" 2 | USER root\n" +
				" 3 | RUN <<'SCRIPT'\n" +
				" 4 | set -ex\n" +
				" 5 | curl -fsSL https://example.com/install.sh | bash\n" +
				" 6 | SCRIPT\n" +
				"\n" +
				"Error: exit code 1\n" +
				"\n" +
				"  + curl -fsSL https://example.com/install.sh | bash\n" +
				"  curl: (6) Could not resolve host: example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &Collector{traces: tt.traces}
			got := collector.Format(tt.dockerfile)
			if got != tt.want {
				t.Errorf("Format() mismatch\nGot:\n%q\nWant:\n%q", got, tt.want)
			}
		})
	}
}

func TestCollector_Format_Integration(t *testing.T) {
	// Test a realistic scenario matching the design doc example
	collector := NewCollector()

	// Simulate receiving BuildKit traces
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

	resp2 := &controlapi.StatusResponse{
		Logs: []*controlapi.VertexLog{
			{Msg: []byte("+ curl -fsSL https://example.com/install.sh | bash\n")},
			{Msg: []byte("curl: (6) Could not resolve host: example.com\n")},
		},
	}
	aux2, _ := createTraceAux(resp2)
	_ = collector.Add(aux2)

	resp3 := &controlapi.StatusResponse{
		Vertexes: []*controlapi.Vertex{
			{
				Digest: "sha256:step1",
				Error:  "exit code 1",
			},
		},
	}
	aux3, _ := createTraceAux(resp3)
	_ = collector.Add(aux3)

	dockerfile := "FROM agentbox/base:0.1.0\nUSER root\nRUN <<'SCRIPT'\nset -ex\ncurl -fsSL https://example.com/install.sh | bash\nSCRIPT"

	result := collector.Format(dockerfile)

	// Verify key components are present
	if !strings.Contains(result, " 1 | FROM agentbox/base:0.1.0") {
		t.Error("Missing FROM line")
	}
	if !strings.Contains(result, "Error: exit code 1") {
		t.Error("Missing error message")
	}
	if !strings.Contains(result, "  + curl -fsSL https://example.com/install.sh | bash") {
		t.Error("Missing build output")
	}
	if !strings.Contains(result, "  curl: (6) Could not resolve host: example.com") {
		t.Error("Missing error output line")
	}
	if strings.Contains(result, "... (") {
		t.Error("Should not have truncation hint for short output")
	}
}
