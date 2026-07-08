package test

import (
	"testing"
)

func TestHelp(t *testing.T) {
	_, err := executeCommand(t, "help")
	if err != nil {
		t.Fatalf("composefork help: %v", err)
	}
}

func TestFoo(t *testing.T) {
	_, err := executeCommand(t, "worktree", "up")
	if err != nil {
		t.Fatalf("composefork help: %v", err)
	}
}
