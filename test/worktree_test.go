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

// TestForkUp brings the project up from a real linked git worktree — the fork
// path, where it runs under its own compose project name rather than the parent's.
func TestForkUp(t *testing.T) {
	project, fork := setupWorktreeTest(t)

	_, err := executeCommand(t, "worktree", "up")
	assertNoError(t, err)
	assertContainerRunning(t, project, "web")
	assertContainerRunning(t, project, "db")
	assertServiceHealthy(t, project, "web")

	// The fork runs under its own project name (parent-feature), not the parent's.
	out, err := executeCommand(t, "ls")
	assertNoError(t, err)
	assertContains(t, out, fork)

	_, err = executeCommand(t, "worktree", "down")
	assertNoError(t, err)
	assertNoContainers(t, project)
}

// TestParallelForks brings up two linked worktrees so they run at the same time,
// verifying forks are isolated: each gets its own compose project name and, with
// published ports stripped, they don't collide on host ports. The app reads its
// project from the process working dir, so the two are brought up one after the
// other — but both stay up together.
func TestParallelForks(t *testing.T) {
	parent, forks := setupForksTest(t, "feature-a", "feature-b")

	for _, f := range forks {
		t.Chdir(f.dir)
		_, err := executeCommand(t, "worktree", "up")
		assertNoError(t, err)
	}

	// Both forks are up at once, each under its own project name.
	out, err := executeCommand(t, "ls")
	assertNoError(t, err)
	assertContains(t, out, forks[0].project)
	assertContains(t, out, forks[1].project)
	assertServiceHealthy(t, parent, "web") // matches both forks' web containers

	for _, f := range forks {
		t.Chdir(f.dir)
		_, err := executeCommand(t, "worktree", "down")
		assertNoError(t, err)
	}
	assertNoContainers(t, parent)
}
