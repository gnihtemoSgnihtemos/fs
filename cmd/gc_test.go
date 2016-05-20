package cmd

import "testing"

func TestDifference(t *testing.T) {
	a := []string{"foo", "bar", "baz"}
	b := []string{"foo"}
	diff := difference(a, b)
	if got := len(diff); got != 2 {
		t.Fatalf("Expected 2 sites, got %d", got)
	}
	if want := "bar"; diff[0] != want {
		t.Errorf("diff[0] = %q, want %q", diff[0], want)
	}
	if want := "baz"; diff[1] != want {
		t.Errorf("diff[1] = %q, want %q", diff[1], want)
	}
}

func TestGCExecute(t *testing.T) {
	err := (&GC{}).Execute([]string{"foo"})
	if err != errUnexpectedArgs {
		t.Errorf("Expected error: %s", errUnexpectedArgs)
	}
}
