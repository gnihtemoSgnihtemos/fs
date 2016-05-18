package cmd

import "testing"

func TestUpdateExecute(t *testing.T) {
	err := (&Update{}).Execute([]string{"foo"})
	if err != errUnexpectedArgs {
		t.Errorf("Expected error: %s", errUnexpectedArgs)
	}
}
