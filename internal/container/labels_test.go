package container

import "testing"

func TestParseLabels(t *testing.T) {
	got, err := ParseLabels([]string{"a=1", "sweatshop.task=abc-123", "empty=", "k=v=w"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"a": "1", "sweatshop.task": "abc-123", "empty": "", "k": "v=w"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("label %q = %q, want %q", k, got[k], v)
		}
	}

	for _, bad := range []string{"novalue", "=v", "bad key=v", "-a=v", "a.=v", "littlebox=2"} {
		if _, err := ParseLabels([]string{bad}); err == nil {
			t.Errorf("ParseLabels(%q) expected error", bad)
		}
	}
}

func TestValidateContainerName(t *testing.T) {
	for _, ok := range []string{"", "ab", "sweatshop-task_1.x"} {
		if err := ValidateContainerName(ok); err != nil {
			t.Errorf("%q: unexpected error %v", ok, err)
		}
	}
	for _, bad := range []string{"a", "-ab", "a b", "a/b"} {
		if err := ValidateContainerName(bad); err == nil {
			t.Errorf("%q: expected error", bad)
		}
	}
}

func TestContainerLabelsAddsManaged(t *testing.T) {
	got := containerLabels(map[string]string{"x": "y"})
	if got[ManagedLabel] != "1" || got["x"] != "y" {
		t.Errorf("containerLabels = %v", got)
	}
	if got := containerLabels(nil); got[ManagedLabel] != "1" {
		t.Errorf("containerLabels(nil) = %v", got)
	}
}
