package profiles

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed Agentfile.*
var agentfiles embed.FS

// Names returns all available profile names in alphabetical order.
// Profile names are derived from embedded Agentfile.* filenames by stripping the "Agentfile." prefix.
func Names() []string {
	entries, err := fs.ReadDir(agentfiles, ".")
	if err != nil {
		// This should never happen with embedded files, but handle gracefully
		return []string{}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if strings.HasPrefix(filename, "Agentfile.") {
			// Extract profile name by removing "Agentfile." prefix
			profileName := strings.TrimPrefix(filename, "Agentfile.")
			names = append(names, profileName)
		}
	}

	sort.Strings(names)
	return names
}

// Get returns the embedded Agentfile content for the specified profile name.
// Returns (content, true) if the profile exists, or ("", false) if not found.
func Get(name string) (string, bool) {
	filename := "Agentfile." + name
	content, err := fs.ReadFile(agentfiles, filename)
	if err != nil {
		return "", false
	}
	return string(content), true
}
