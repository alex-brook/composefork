package test

import (
	"testing"
)

func TestHelp(t *testing.T) {
	setupTest(t)

	out, err := executeCommand(t, "help")
	assertNoError(t, err)
	assertContains(t, out, "worktree")
	assertNotContains(t, out, "panic")

	// setupTest generated a .env pointing at the devcontainer compose file.
	assertFileExists(t, ".env")
	assertFileContains(t, ".env", "COMPOSE_FILE=.devcontainer/compose.yml")
}
