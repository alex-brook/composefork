package test

import (
	"testing"
)

// Standalone happy-path tests for commands that don't need a running fork.

func TestVersion(t *testing.T) {
	out, err := executeCommand(t, "version")
	assertNoError(t, err)
	assertContains(t, out, "dev") // default version when not built with ldflags
}

func TestSetup(t *testing.T) {
	out, err := executeCommand(t, "setup")
	assertNoError(t, err)
	assertContains(t, out, "composefork")
	assertContains(t, out, "Assumptions composefork makes")
}

// ls is global (lists forks across all worktrees); with no fork up this just
// exercises the command's plumbing. "Shows my fork" is covered in TestWorktree.
func TestLs(t *testing.T) {
	_, err := executeCommand(t, "ls")
	assertNoError(t, err)
}
