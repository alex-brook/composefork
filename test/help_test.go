package test

import (
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	setupTest(t)

	out, err := executeCommand(t, "help")
	assertNoError(t, err)
	assertContains(t, out, "up")
	assertContains(t, out, "restart")
	assertNotContains(t, out, "panic")

	// Root help is the agent entry point: it must carry the usage guidance, not
	// just the command list, and must point at setup for project prerequisites.
	assertContains(t, out, "[FOR AGENTS]")
	assertContains(t, out, "composefork setup")
	assertNotContains(t, out, "Project requirements")

	// The terminal scrolls to the end, so the agent briefing leads and the
	// command list comes last, leaving the commands on screen.
	if strings.Index(out, "[FOR AGENTS]") > strings.Index(out, "Available Commands") {
		t.Fatalf("root help should put the agent briefing before the command list\n--- output ---\n%s", out)
	}

	// setupTest generated a .env pointing at the devcontainer compose file.
	assertFileExists(t, ".env")
	assertFileContains(t, ".env", "COMPOSE_FILE=.devcontainer/compose.yml")
}
