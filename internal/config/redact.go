package config

import (
	"os"
	"path/filepath"
)

// GetRedactedFiles returns a list of absolute file paths that match the redact patterns.
// Patterns support exact filenames (.env) and glob patterns (*.pem, secrets/*).
// Matching is case-sensitive.
func GetRedactedFiles(workspacePath string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	var redacted []string

	err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(workspacePath, path)

		for _, pattern := range patterns {
			// Match against relative path (for patterns like secrets/*)
			matched, _ := filepath.Match(pattern, relPath)
			if matched {
				redacted = append(redacted, path)
				break
			}
			// Match against basename (for patterns like .env, *.pem)
			matched, _ = filepath.Match(pattern, filepath.Base(path))
			if matched {
				redacted = append(redacted, path)
				break
			}
		}

		return nil
	})

	return redacted, err
}
