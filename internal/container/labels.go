package container

import (
	"fmt"
	"regexp"
	"strings"
)

// ManagedLabel is added to every container littlebox creates so they can be
// found with `docker ps --filter label=littlebox=1`.
const ManagedLabel = "littlebox"

// CreateOptions holds optional metadata applied to a created container.
type CreateOptions struct {
	// Name is the container name. Empty lets Docker generate one.
	Name string
	// Labels are user-supplied labels. ManagedLabel=1 is always added.
	Labels map[string]string
}

var (
	labelKeyRe      = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._/-]*[A-Za-z0-9])?$`)
	containerNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)
)

// ParseLabels parses repeated "key=value" strings into a map.
// Keys must be non-empty and use only [A-Za-z0-9._/-], starting and ending
// with an alphanumeric; values may be empty. The ManagedLabel key is reserved.
func ParseLabels(raw []string) (map[string]string, error) {
	labels := make(map[string]string, len(raw))
	for _, kv := range raw {
		key, value, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("invalid label %q: expected key=value", kv)
		}
		if !labelKeyRe.MatchString(key) {
			return nil, fmt.Errorf("invalid label key %q: must match %s", key, labelKeyRe.String())
		}
		if key == ManagedLabel {
			return nil, fmt.Errorf("label key %q is reserved", ManagedLabel)
		}
		labels[key] = value
	}
	return labels, nil
}

// ValidateContainerName checks a name against Docker's container name rules.
// An empty name is valid (Docker generates one).
func ValidateContainerName(name string) error {
	if name == "" || containerNameRe.MatchString(name) {
		return nil
	}
	return fmt.Errorf("invalid container name %q: must match %s", name, containerNameRe.String())
}

// containerLabels returns the user labels merged with ManagedLabel=1.
func containerLabels(user map[string]string) map[string]string {
	labels := make(map[string]string, len(user)+1)
	for k, v := range user {
		labels[k] = v
	}
	labels[ManagedLabel] = "1"
	return labels
}
