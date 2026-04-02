package cmd

import (
	"testing"
)

func TestNewRootCmdContainsExpectedCommands(t *testing.T) {
	root := NewRootCmd()

	expected := []string{"secrets", "keys", "certificates", "get", "set", "rm", "ensure", "use", "version"}
	for _, name := range expected {
		if root.CommandPath() == name {
			t.Fatalf("unexpected command path collision for %q", name)
		}

		if cmd, _, err := root.Find([]string{name}); err != nil || cmd == nil {
			t.Fatalf("expected command %q to exist", name)
		}
	}
}
