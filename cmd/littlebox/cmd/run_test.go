package cmd

import "testing"

func TestRunFlagsLabelAndName(t *testing.T) {
	f := runCmd.Flags()
	if f.Lookup("name") == nil {
		t.Fatal("--name flag not registered")
	}
	if err := f.Parse([]string{"--label", "a=1", "--label", "b=2", "--name", "box"}); err != nil {
		t.Fatal(err)
	}
	got, _ := f.GetStringArray("label")
	if len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
		t.Errorf("labels = %v", got)
	}
	if name != "box" {
		t.Errorf("name = %q", name)
	}
	labels, name = nil, ""
}
