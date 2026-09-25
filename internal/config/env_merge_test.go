package config

import (
	"reflect"
	"testing"
)

func TestMergeEnv(t *testing.T) {
	got := MergeEnv(
		[]string{"A=pass", "B=pass", "C=x=y"},
		[]EnvVar{{Name: "B", Value: "top"}, {Name: "D", Value: "top"}},
		[]EnvVar{{Name: "D", Value: "agent"}, {Name: "E", Value: ""}, {Name: "A", Value: "agent"}},
	)
	want := []string{"A=agent", "B=top", "C=x=y", "D=agent", "E="}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMergeEnvEmpty(t *testing.T) {
	if got := MergeEnv(nil, nil, nil); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestValidateEnvVars(t *testing.T) {
	good := []EnvVar{{Name: "GIT_AUTHOR_NAME"}, {Name: "_x1"}}
	if errs := validateEnvVars("agent.env", good); len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	for _, n := range []string{"", "1A", "A-B", "A B", "A=B"} {
		if errs := validateEnvVars("env", []EnvVar{{Name: n}}); len(errs) != 1 {
			t.Errorf("name %q: expected 1 error, got %v", n, errs)
		}
	}
}
