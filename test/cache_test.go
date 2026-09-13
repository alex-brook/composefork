package test

import (
	"testing"
)

// forkEdit stands in for the work a fork does after it starts: whatever the
// agent installed, migrated or wrote since the snapshot was taken.
const forkEdit = "composefork-fork-edit"

// TestCache exercises the with-cache path: snapshot the project's volumes with
// `cache`, then a fork imports the snapshot on `up`. Mirrors the intended
// workflow — cache once, then start forks faster from it. Like TestWorktree it
// shares one expensive setup across ordered subtests, so they are not
// independent.
func TestCache(t *testing.T) {
	dir, project := newRepo(t)
	populateProject(t, dir, project)
	t.Chdir(dir)

	// Run the parent the way a developer would before caching, leaving its volumes
	// in place.
	primeParent(t)

	_, err := executeCommand(t, "cache")
	assertNoError(t, err)
	// A primed snapshot holds the installed gems (tens of MB); an empty/wrong
	// volume would be a few KB, so this guards against snapshotting nothing.
	if info := statCacheTarball(t, project+"_bundle_data.tar"); info.Size() < 1<<20 {
		t.Fatalf("cache tarball is %d bytes; expected a primed (non-empty) snapshot", info.Size())
	}

	// The cache is for forks, so the consumer has to be a real linked worktree —
	// in the main checkout the volumes already exist and are never imported into.
	wt := addWorktree(t, dir, "feature")
	populateProject(t, wt, project)
	t.Chdir(wt)
	bundleVolume := project + "-feature_bundle_data"

	t.Run("up restores the snapshot", func(t *testing.T) {
		out, err := executeCommand(t, "up")
		assertNoError(t, err)
		assertContains(t, out, "Restoring cached volumes")
		assertServiceHealthy(t, project, "web")
	})

	t.Run("up again keeps the fork's own data", func(t *testing.T) {
		path := gemFile(t, bundleVolume)
		writeVolumeFile(t, bundleVolume, path, forkEdit)

		out, err := executeCommand(t, "up")
		assertNoError(t, err)
		assertNotContains(t, out, "Restoring cached volumes")

		// Truncated: restoring puts the gem's own contents back, which are long
		if got := readVolumeFile(t, bundleVolume, path); got != forkEdit {
			t.Fatalf("%s in %s is %.60q after a second up, want %q: the snapshot was restored over the fork's own data",
				path, bundleVolume, got, forkEdit)
		}
	})

	_, err = executeCommand(t, "down")
	assertNoError(t, err)
	assertNoContainers(t, project)
}
