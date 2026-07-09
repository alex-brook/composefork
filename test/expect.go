package test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex-brook/composefork/internal"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
)

// Expectations for asserting things about composefork's behavior. Every helper
// is fatal (t.Fatalf) and calls t.Helper() so failures point at the caller.
//
// Docker/service helpers scope to a single test via the composefork.project
// label, whose value is the project name returned by setupTest.

// --- Command result -------------------------------------------------------

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func assertContains(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Fatalf("output missing %q\n--- output ---\n%s", want, out)
	}
}

func assertNotContains(t *testing.T, out, notWant string) {
	t.Helper()
	if strings.Contains(out, notWant) {
		t.Fatalf("output unexpectedly contains %q\n--- output ---\n%s", notWant, out)
	}
}

// --- Docker resources -----------------------------------------------------

// forkContainers lists containers labeled for the given fork project. A
// non-empty service narrows results to that single compose service.
func forkContainers(t *testing.T, project, service string) []container.Summary {
	t.Helper()
	f := make(client.Filters).Add("label", internal.COMPOSEFORK_PROJECT_LABEL+"="+project)
	if service != "" {
		f.Add("label", api.ServiceLabel+"="+service)
	}
	list, err := dockerClient(t).ContainerList(context.Background(), client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		t.Fatalf("listing containers for project %q: %v", project, err)
	}
	return list.Items
}

func assertContainerExists(t *testing.T, project, service string) {
	t.Helper()
	if len(forkContainers(t, project, service)) == 0 {
		t.Fatalf("expected a container for service %q in project %q, found none", service, project)
	}
}

func assertContainerRunning(t *testing.T, project, service string) {
	t.Helper()
	cs := forkContainers(t, project, service)
	if len(cs) == 0 {
		t.Fatalf("expected a running container for service %q in project %q, found none", service, project)
	}
	for _, c := range cs {
		if c.State != container.StateRunning {
			t.Fatalf("service %q container %s is %q, want running", service, c.ID, c.State)
		}
	}
}

func assertNoContainers(t *testing.T, project string) {
	t.Helper()
	if cs := forkContainers(t, project, ""); len(cs) != 0 {
		t.Fatalf("expected no containers for project %q, found %d", project, len(cs))
	}
}

func assertNetworkExists(t *testing.T, project, name string) {
	t.Helper()
	f := make(client.Filters).
		Add("label", internal.COMPOSEFORK_PROJECT_LABEL+"="+project).
		Add("label", api.NetworkLabel+"="+name)
	list, err := dockerClient(t).NetworkList(context.Background(), client.NetworkListOptions{Filters: f})
	if err != nil {
		t.Fatalf("listing networks for project %q: %v", project, err)
	}
	if len(list.Items) == 0 {
		t.Fatalf("expected network %q in project %q, found none", name, project)
	}
}

func forkVolumes(t *testing.T, project, name string) []volume.Volume {
	t.Helper()
	f := make(client.Filters).
		Add("label", internal.COMPOSEFORK_PROJECT_LABEL+"="+project).
		Add("label", api.VolumeLabel+"="+name)
	list, err := dockerClient(t).VolumeList(context.Background(), client.VolumeListOptions{Filters: f})
	if err != nil {
		t.Fatalf("listing volumes for project %q: %v", project, err)
	}
	return list.Items
}

func assertVolumeExists(t *testing.T, project, name string) {
	t.Helper()
	if len(forkVolumes(t, project, name)) == 0 {
		t.Fatalf("expected volume %q in project %q, found none", name, project)
	}
}

func assertVolumeGone(t *testing.T, project, name string) {
	t.Helper()
	if len(forkVolumes(t, project, name)) != 0 {
		t.Fatalf("expected volume %q in project %q to be gone, still present", name, project)
	}
}

// --- Service health -------------------------------------------------------

func assertServiceHealthy(t *testing.T, project, service string) {
	t.Helper()
	cs := forkContainers(t, project, service)
	if len(cs) == 0 {
		t.Fatalf("expected a container for service %q in project %q, found none", service, project)
	}
	// Health must come from inspect: Summary.Health (from the list endpoint) is
	// omitempty and comes back nil on most daemons even when a healthcheck exists.
	for _, c := range cs {
		res, err := dockerClient(t).ContainerInspect(context.Background(), c.ID, client.ContainerInspectOptions{})
		if err != nil {
			t.Fatalf("inspecting service %q container %s: %v", service, c.ID, err)
		}
		if res.Container.State == nil || res.Container.State.Health == nil {
			t.Fatalf("service %q container %s has no healthcheck", service, c.ID)
		}
		if status := res.Container.State.Health.Status; status != container.Healthy {
			t.Fatalf("service %q container %s health is %q, want healthy", service, c.ID, status)
		}
	}
}

// --- Filesystem / cache ---------------------------------------------------

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %q: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %q: %v", path, err)
	}
	if !strings.Contains(string(b), want) {
		t.Fatalf("file %q missing %q\n--- contents ---\n%s", path, want, b)
	}
}

// statCacheTarball asserts a volume snapshot exists in composefork's cache dir
// (where `cache` writes tarballs consumed by importVolumes on up) and returns
// its FileInfo so callers can also check size.
func statCacheTarball(t *testing.T, name string) os.FileInfo {
	t.Helper()
	dir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("resolving user cache dir: %v", err)
	}
	path := filepath.Join(dir, internal.APP_NAME, name)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected cache tarball %q: %v", path, err)
	}
	return info
}
