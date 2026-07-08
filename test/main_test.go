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

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/moby/moby/client"
)

const dotEnvTemplate = `COMPOSE_FILE=.devcontainer/compose.yml
COMPOSE_PROJECT_NAME={{.ProjectName}}
`

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	// Setup temp dir
	dir, err := os.MkdirTemp("tmp", "test-*")
	if err != nil {
		log.Fatalf("tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

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

	// Tidy up docker stuff created by the test run
	docker, err := command.NewDockerCli()
	if err != nil {
		log.Fatalf("error creating docker cli: %v", err)
	}
	if err := docker.Initialize(&flags.ClientOptions{}, command.WithOutputStream(os.Stdout)); err != nil {
		log.Fatalf("error initializing docker cli: %v", err)
	}
	defer func() {
		err := cleanupRun(docker.Client(), runName)
		if err != nil {
			log.Fatalf("error tearing down test: %v", err)
		}
	}()

	// Enter the directory for this test run
	oldWd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error entering test dir: %v", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Fatalf("error entering test dir: %v", err)
	}
	defer os.Chdir(oldWd)

	// Start a service gemstash container

	return m.Run()
}

func cleanupRun(docker client.APIClient, name string) error {
	fmt.Println("cleaning up", name)
	ctx := context.Background()
	f := make(client.Filters).Add("label", fmt.Sprintf("com.github.alex-brook.composefork.project=%s", name))
	var errs []error

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

	// Images are shared between forks, that's why we're using the top level filter here
	imgF := make(client.Filters).Add("label", fmt.Sprintf("com.docker.compose.project=%s", name))
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

	return errors.Join(errs...)
}
