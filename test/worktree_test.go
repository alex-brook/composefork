package test

import (
	"testing"
)

// TestWorktree is the ordered end-to-end lifecycle for the no-cache path. It
// shares one `up` across subtests (up is the expensive step), so the subtests
// run in order and are not independent. The with-cache path lives in TestCache.
func TestWorktree(t *testing.T) {
	project := setupTest(t)

	t.Run("up cold", func(t *testing.T) {
		_, err := executeCommand(t, "worktree", "up")
		assertNoError(t, err)

		assertContainerRunning(t, project, "web")
		assertContainerRunning(t, project, "db")
		assertServiceHealthy(t, project, "web")
		assertNetworkExists(t, project, "default")
		assertVolumeExists(t, project, "bundle_data")
	})

	t.Run("ps", func(t *testing.T) {
		_, err := executeCommand(t, "worktree", "ps")
		assertNoError(t, err)
	})

	t.Run("ls", func(t *testing.T) {
		out, err := executeCommand(t, "ls")
		assertNoError(t, err)
		assertContains(t, out, project)
	})

	t.Run("exec", func(t *testing.T) {
		_, err := executeCommand(t, "worktree", "exec", "web", "true")
		assertNoError(t, err)
	})

	t.Run("restart", func(t *testing.T) {
		_, err := executeCommand(t, "worktree", "restart")
		assertNoError(t, err)
		// Running again; not asserting healthy — restart resets the 90s health
		// start period, so healthy here would flake.
		assertContainerRunning(t, project, "web")
	})

	t.Run("down", func(t *testing.T) {
		_, err := executeCommand(t, "worktree", "down")
		assertNoError(t, err)
		assertNoContainers(t, project)
	})
}
