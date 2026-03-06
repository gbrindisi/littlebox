package builderror

import (
	"fmt"
	"strings"
)

// Format returns a formatted error message with Dockerfile context and build output.
// It shows the Dockerfile with line numbers, highlights the failed line, displays
// the error message, and includes truncated build output (last 20 lines).
func (c *Collector) Format(dockerfileContent string) string {
	var result strings.Builder

	// Format Dockerfile section
	dockerfileSection := formatDockerfile(dockerfileContent, 0) // We'll implement line detection later
	result.WriteString(dockerfileSection)
	result.WriteString("\n")

	// Format error message section
	errorMsg := c.GetError()
	if errorMsg != "" {
		result.WriteString(fmt.Sprintf("Error: %s\n", errorMsg))
		result.WriteString("\n")
	}

	// Format build output section
	logs := c.GetLogs()
	outputSection := formatOutput(logs)
	result.WriteString(outputSection)

	return result.String()
}

// formatDockerfile formats the Dockerfile with line numbers and highlights the failed line.
// Line numbers are formatted as "  N | <line>" with failed lines using "> N | <line>".
func formatDockerfile(content string, failedLine int) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i, line := range lines {
		lineNum := i + 1
		if lineNum == failedLine {
			result.WriteString(fmt.Sprintf("> %d | %s\n", lineNum, line))
		} else {
			result.WriteString(fmt.Sprintf(" %d | %s\n", lineNum, line))
		}
	}

	return result.String()
}

// formatOutput formats the build output with truncation to the last 20 lines.
// If output exceeds 20 lines, it shows a hint message with the count of hidden lines.
func formatOutput(logs []string) string {
	const maxLines = 20

	if len(logs) == 0 {
		return ""
	}

	var result strings.Builder

	if len(logs) > maxLines {
		// Show only last 20 lines
		truncatedLogs := logs[len(logs)-maxLines:]
		for _, log := range truncatedLogs {
			result.WriteString(fmt.Sprintf("  %s\n", log))
		}
		// Add truncation hint
		hiddenLines := len(logs) - maxLines
		result.WriteString(fmt.Sprintf("\n... (%d more lines, use --debug for full output)\n", hiddenLines))
	} else {
		// Show all lines
		for _, log := range logs {
			result.WriteString(fmt.Sprintf("  %s\n", log))
		}
	}

	return result.String()
}
