package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/alex-brook/composefork/cmd"
	"github.com/alex-brook/composefork/internal"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/client"
)

// primeParent brings the parent compose project up (building as needed) and
// waits for it to become healthy, populating its volumes the way a normal
// `docker compose up` would, then stops it leaving the volumes in place for
// `cache` to snapshot. Registers cleanup that removes the parent and its
// volumes. Uses the same compose library the app itself does.
func primeParent(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	app, err := internal.NewApp()
	if err != nil {
		t.Fatalf("creating app: %v", err)
	}
	fork, err := internal.LoadFork()
	if err != nil {
		t.Fatalf("loading parent project: %v", err)
	}
	parent := fork.Parent

	down := func() error {
		return app.Compose.Down(ctx, parent.Name, api.DownOptions{Volumes: true, RemoveOrphans: true})
	}
	// The parent project name is reused across runs and its containers carry no
	// composefork label (so setupTest's cleanup won't catch them). Clear any
	// leftovers up front, and tear down on the way out.
	_ = down()
	t.Cleanup(func() {
		if err := down(); err != nil {
			t.Errorf("tearing down parent: %v", err)
		}
	})

	if err := app.Compose.Build(ctx, parent, api.BuildOptions{}); err != nil {
		t.Fatalf("building parent: %v", err)
	}
	// Up with Create creates and starts in one call (production splits them only
	// to seed volumes in between, which priming doesn't need).
	if err := app.Compose.Up(ctx, parent, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{Wait: true}}); err != nil {
		t.Fatalf("bringing up parent: %v", err)
	}
	// Stop the parent but keep its now-primed volumes for `cache` to snapshot.
	if err := app.Compose.Down(ctx, parent.Name, api.DownOptions{Volumes: false, RemoveOrphans: true}); err != nil {
		t.Fatalf("stopping parent: %v", err)
	}
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

// setupTest prepares an isolated project in a temp dir and registers Docker
// cleanup. It returns the project name (the composefork.project label value)
// so assertions can be scoped to this test's resources.
func setupTest(t *testing.T) string {
	t.Helper()

	// Isolate the cache dir (os.UserCacheDir honors XDG_CACHE_HOME on Linux) so
	// cold-up sees no cache and the cache command never touches the real ~/.cache.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// Setup temp dir
	dir := t.TempDir()

	// Copy the dummy project contents into our test project
	if err := os.CopyFS(dir, os.DirFS("dummy")); err != nil {
		log.Fatalf("couldn't copy dummy: %v", err)
	}

	// Create a .env that points to the devcontainer
	runName := strings.Split(dir, "/")[1]
	var buf bytes.Buffer
	template.
		Must(template.New("dotenv").Parse(dotEnvTemplate)).
		Execute(&buf, struct{ ProjectName string }{ProjectName: runName})
	if err := os.WriteFile(filepath.Join(dir, ".env"), buf.Bytes(), 0644); err != nil {
		log.Fatalf("couldn't create .env: %v", err)
	}

	t.Chdir(dir)

	t.Cleanup(func() { cleanupDocker(t, runName) })

	return runName
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
