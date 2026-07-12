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
		_, err := executeCommand(t, "up")
		assertNoError(t, err)

		assertContainerRunning(t, project, "web")
		assertContainerRunning(t, project, "db")
		assertServiceHealthy(t, project, "web")
		assertNetworkExists(t, project, "default")
		assertVolumeExists(t, project, "bundle_data")
	})

	t.Run("ps", func(t *testing.T) {
		_, err := executeCommand(t, "ps")
		assertNoError(t, err)
	})

	t.Run("ls", func(t *testing.T) {
		out, err := executeCommand(t, "ls")
		assertNoError(t, err)
		assertContains(t, out, project)
	})

	t.Run("exec", func(t *testing.T) {
		_, err := executeCommand(t, "exec", "web", "true")
		assertNoError(t, err)
	})

	t.Run("restart", func(t *testing.T) {
		_, err := executeCommand(t, "restart")
		assertNoError(t, err)
		// Running again; not asserting healthy — restart resets the 90s health
		// start period, so healthy here would flake.
		assertContainerRunning(t, project, "web")
	})

	t.Run("down", func(t *testing.T) {
		_, err := executeCommand(t, "down")
		assertNoError(t, err)
		assertNoContainers(t, project)
	})
}

// TestExecCapturesOutput proves command output flows through the writer threaded
// into the app rather than escaping to os.Stdout: it execs a command in a service
// and finds its stdout in the captured buffer. It also confirms the app's log
// diagnostics ("Building") are captured on the up call.
func TestExecCapturesOutput(t *testing.T) {
	project := setupTest(t)

	out, err := executeCommand(t, "up")
	assertNoError(t, err)
	assertServiceHealthy(t, project, "web")
	assertContains(t, out, "Building") // a.Log diagnostic reached the writer

	out, err = executeCommand(t, "exec", "web", "sh", "-c", "echo composefork-marker")
	assertNoError(t, err)
	assertContains(t, out, "composefork-marker")

	_, err = executeCommand(t, "down")
	assertNoError(t, err)
	assertNoContainers(t, project)
}

// TestForkUp brings the project up from a real linked git worktree — the fork
// path, where it runs under its own compose project name rather than the parent's.
func TestForkUp(t *testing.T) {
	project, fork := setupWorktreeTest(t)

	_, err := executeCommand(t, "up")
	assertNoError(t, err)
	assertContainerRunning(t, project, "web")
	assertContainerRunning(t, project, "db")
	assertServiceHealthy(t, project, "web")

	// The fork runs under its own project name (parent-feature), not the parent's.
	out, err := executeCommand(t, "ls")
	assertNoError(t, err)
	assertContains(t, out, fork)

	_, err = executeCommand(t, "down")
	assertNoError(t, err)
	assertNoContainers(t, project)
}

// TestForkBuildsOwnImage proves a child worktree builds its own image from its
// own modified source: it appends a unique marker to the fork's Dockerfile,
// brings the fork up, and checks the image Compose built for that fork carries
// the marker.
func TestForkBuildsOwnImage(t *testing.T) {
	parent, forks := setupForksTest(t, "feature")
	f := forks[0]
	marker := f.project // unique to this fork

	bakeMarker(t, f.dir, marker)

	t.Chdir(f.dir)
	_, err := executeCommand(t, "up")
	assertNoError(t, err)
	assertServiceHealthy(t, parent, "web")

	// The fork built its own image from its modified Dockerfile.
	assertImageEnv(t, f.project, "COMPOSEFORK_MARKER="+marker)

	_, err = executeCommand(t, "down")
	assertNoError(t, err)
	assertNoContainers(t, parent)
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
		_, err := executeCommand(t, "up")
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
		_, err := executeCommand(t, "down")
		assertNoError(t, err)
	}
	assertNoContainers(t, parent)
}
