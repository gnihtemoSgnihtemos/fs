package cmd

import "testing"

func TestTestExecute(t *testing.T) {
	err := (&Test{}).Execute([]string{"foo"})
	if err != errUnexpectedArgs {
		t.Errorf("Expected error: %s", errUnexpectedArgs)
	}
}
