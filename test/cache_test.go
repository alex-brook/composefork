package test

import (
	"testing"
)

// TestCache exercises the with-cache path: prime the parent project's volumes by
// running it normally (docker compose up), snapshot them with `cache`, then a
// fork imports the snapshot on `up`. Mirrors the intended workflow —
// cache the primed parent, then start forks faster from it.
func TestCache(t *testing.T) {
	project := setupTest(t)

	// Prime the parent's volumes by running it normally, leaving them in place
	// for `cache` to snapshot.
	primeParent(t)

	_, err := executeCommand(t, "cache")
	assertNoError(t, err)
	// A primed snapshot holds the installed gems (tens of MB); an empty/wrong
	// volume would be a few KB, so this guards against snapshotting nothing.
	if info := statCacheTarball(t, project+"_bundle_data.tar"); info.Size() < 1<<20 {
		t.Fatalf("cache tarball is %d bytes; expected a primed (non-empty) snapshot", info.Size())
	}

	// A fork now starts from the cached volumes and comes up healthy.
	_, err = executeCommand(t, "up")
	assertNoError(t, err)
	assertServiceHealthy(t, project, "web")

	_, err = executeCommand(t, "down")
	assertNoError(t, err)
	assertNoContainers(t, project)
}
