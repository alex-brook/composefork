package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/alex-brook/composefork/cmd"
	"github.com/alex-brook/composefork/internal"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/moby/moby/client"
)

// primeParent brings the parent project up and waits for health so its volumes
// are populated the way a normal `docker compose up` would, then stops it
// leaving the volumes for `cache` to snapshot. Registers cleanup that removes
// the parent and its volumes.
//
// This shells out to `docker compose` (rather than the compose library used
// elsewhere) because the library can't run this authored, build-based parent
// project the way the CLI does — the CLI handles build→image resolution and
// depends_on sequencing that the library only does for the app's already
// image-resolved fork projects.
func primeParent(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { composeParent(t, "down", "-v") })
	composeParent(t, "up", "-d", "--wait")
	composeParent(t, "down") // remove containers, keep the primed volumes
}

// composeParent runs `docker compose <args>` against the parent project in the
// test's working dir (using the .env setupTest wrote).
func composeParent(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker compose %v: %v", args, err)
	}
}

// runGit runs a git command in dir, wiring output through and failing the test
// on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
}

// initGitRepo makes dir a git repository so the app's main-worktree detection
// has a repo to inspect.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
}

// dockerClient builds a Docker API client for querying/cleaning up test
// resources. Fatal on failure — every caller needs a working daemon.
func dockerClient(t *testing.T) client.APIClient {
	t.Helper()
	cli, err := command.NewDockerCli()
	if err != nil {
		t.Fatalf("error creating docker cli: %v", err)
	}
	if err := cli.Initialize(&flags.ClientOptions{}, command.WithOutputStream(os.Stdout)); err != nil {
		t.Fatalf("error initializing docker cli: %v", err)
	}
	return cli.Client()
}

const dotEnvTemplate = `COMPOSE_FILE=.devcontainer/compose.yml
COMPOSE_PROJECT_NAME={{.ProjectName}}
`

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	root := cmd.NewRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// newRepo makes an isolated git repo in a temp dir and registers Docker cleanup,
// returning the repo dir and the derived project (parent) name. It does not copy
// the project files or change directory — callers do that for whichever working
// dir (main checkout or linked worktree) the test exercises.
func newRepo(t *testing.T) (dir, project string) {
	t.Helper()

	// Isolate the cache dir (os.UserCacheDir honors XDG_CACHE_HOME on Linux) so
	// cold-up sees no cache and the cache command never touches the real ~/.cache.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// The app detects the main checkout vs. a linked worktree via git, so the
	// project needs a real repo to run in.
	dir = t.TempDir()
	initGitRepo(t, dir)

	// Derive a project name from t.TempDir's unique-per-run parent segment (the
	// leaf is only a per-call counter), lowercased for compose. This isolates each
	// run — a shared name lets an interrupted run's leftovers collide with (and
	// poison) later runs.
	project = strings.ToLower(filepath.Base(filepath.Dir(dir)))

	t.Cleanup(func() { cleanupDocker(t, project) })

	return dir, project
}

// populateProject copies the dummy compose project into dir and writes a .env
// pointing at it with the given project name.
func populateProject(t *testing.T, dir, project string) {
	t.Helper()

	if err := os.CopyFS(dir, os.DirFS("dummy")); err != nil {
		log.Fatalf("couldn't copy dummy: %v", err)
	}

	var buf bytes.Buffer
	template.
		Must(template.New("dotenv").Parse(dotEnvTemplate)).
		Execute(&buf, struct{ ProjectName string }{ProjectName: project})
	if err := os.WriteFile(filepath.Join(dir, ".env"), buf.Bytes(), 0644); err != nil {
		log.Fatalf("couldn't create .env: %v", err)
	}
}

// addWorktree creates a linked git worktree named `name` off the repo at repoDir
// and returns its path. A worktree checks out a commit, so the repo needs a HEAD;
// an empty one is created the first time (the checkout is empty — populateProject
// writes the project files afterward). Distinct names give isolated worktrees:
// separate branches and, via forkName, separate compose project names.
func addWorktree(t *testing.T, repoDir, name string) string {
	t.Helper()

	if exec.Command("git", "-C", repoDir, "rev-parse", "--verify", "-q", "HEAD").Run() != nil {
		runGit(t, repoDir,
			"-c", "user.email=test@example.com", "-c", "user.name=test",
			"commit", "--allow-empty", "-m", "init")
	}

	wt := filepath.Join(t.TempDir(), name)
	runGit(t, repoDir, "worktree", "add", wt)

	return wt
}

// bakeMarker appends an instruction to the worktree's Dockerfile that bakes a
// unique marker into the built image, so a test can prove the image was rebuilt
// from this worktree's own source. Appended last so prior build layers stay
// cached.
func bakeMarker(t *testing.T, dir, marker string) {
	t.Helper()
	path := filepath.Join(dir, ".devcontainer", "Dockerfile")
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("opening %q: %v", path, err)
	}
	defer fh.Close()
	if _, err := fmt.Fprintf(fh, "\nENV COMPOSEFORK_MARKER=%s\n", marker); err != nil {
		t.Fatalf("writing marker to %q: %v", path, err)
	}
}

// fork is a linked worktree and the compose project name it runs under.
type fork struct {
	dir     string
	project string
}

// setupForksTest prepares an isolated repo with one linked worktree per name,
// each populated as a fork. It returns the parent project name (the label value
// assertions filter on) and the forks. It does not change directory — callers
// chdir into a fork before driving the app there.
func setupForksTest(t *testing.T, names ...string) (parent string, forks []fork) {
	t.Helper()

	dir, parent := newRepo(t)
	for _, name := range names {
		wt := addWorktree(t, dir, name)
		populateProject(t, wt, parent)
		forks = append(forks, fork{dir: wt, project: parent + "-" + name})
	}

	return parent, forks
}

// setupTest prepares an isolated project in the main worktree and registers
// Docker cleanup. It returns the project name (the composefork.project label
// value) so assertions can be scoped to this test's resources.
func setupTest(t *testing.T) string {
	t.Helper()

	dir, project := newRepo(t)
	populateProject(t, dir, project)
	t.Chdir(dir)

	return project
}

// setupWorktreeTest prepares an isolated project in a single real linked git
// worktree, exercising the fork path, and chdirs into it. It returns the parent
// project name (the label value assertions filter on) and the fork's own compose
// project name.
func setupWorktreeTest(t *testing.T) (project, fork string) {
	t.Helper()

	parent, forks := setupForksTest(t, "feature")
	t.Chdir(forks[0].dir)

	return parent, forks[0].project
}

func cleanupDocker(t *testing.T, projectName string) {
	t.Helper()
	fmt.Println("cleaning up", projectName)
	ctx := context.Background()
	f := make(client.Filters).Add("label", fmt.Sprintf("%s=%s", internal.COMPOSEFORK_PROJECT_LABEL, projectName))
	var errs []error

	docker := dockerClient(t)

	// 1. Containers — force-kill running ones, remove their anon volumes.
	//    All:true so running AND stopped are included.
	cList, err := docker.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, c := range cList.Items {
			if _, err := docker.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{
				Force:         true, // kill if running
				RemoveVolumes: true,
			}); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 2. Networks — must come after the containers that were attached to them.
	nList, err := docker.NetworkList(ctx, client.NetworkListOptions{Filters: f})
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, n := range nList.Items {
			if _, err := docker.NetworkRemove(ctx, n.ID, client.NetworkRemoveOptions{}); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 3. Volumes — force removes even if referenced/in-use.
	vList, err := docker.VolumeList(ctx, client.VolumeListOptions{Filters: f})
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, v := range vList.Items {
			if _, err := docker.VolumeRemove(ctx, v.Name, client.VolumeRemoveOptions{Force: true}); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Images are shared between forks, so match on Compose's own project label
	// rather than ours.
	imgF := make(client.Filters).Add("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))
	iList, err := docker.ImageList(ctx, client.ImageListOptions{All: true, Filters: imgF})
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, img := range iList.Items {
			if _, err := docker.ImageRemove(ctx, img.ID, client.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			}); err != nil {
				errs = append(errs, err)
			}
		}
	}

	joinedErr := errors.Join(errs...)
	if errs != nil {
		log.Fatalf("teardown error: %v", joinedErr)
	}
}
