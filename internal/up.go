package internal

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (a *App) Up() error {
	// Resolve parent / master project
	log.Println("Loading parent project")
	fork, err := LoadFork()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	// Build the master project
	log.Println("Building", fork.Parent.Name)
	err = a.Compose.Build(context.Background(), fork.Parent, api.BuildOptions{})
	if err != nil {
		return fmt.Errorf("error building project: %w", err)
	}

	// Create a child project for the worktree
	log.Println("Creating fork project")
	project, err := fork.Project()
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	err = a.Compose.Create(context.Background(), project, api.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	err = a.importVolumes(fork, project)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	// Bring a child project up
	log.Println("Bringing up project", project.Name)
	err = a.Compose.Up(context.Background(), project, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{Wait: true}})
	if err != nil {
		return fmt.Errorf("up error: %w", err)
	}

	// Print project info
	return a.printProjectStatus(project.Name)
}

func (a *App) importVolumes(fork *Fork, project *types.Project) error {
	dir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("error copying volumes: %w", err)
	}
	dir = filepath.Join(dir, APP_NAME)

	binds := []string{fmt.Sprintf("%s:/cache", dir)}
	commands := []string{"sh", "-c", `
    set -e
    echo "all args $@"
    while [ "$#" -ge 2 ]; do
      echo "copying $1 to $2"
      tar --numeric-owner -xpf "$1" -C "$2" --strip-components=1
      ls -la $2
      shift 2
    done
  `, "foo"}
	for _, vol := range project.Volumes {
		// set up the commands the container will run to copy the cached
		// snapshots into the child volumes
		expectedTarball := fmt.Sprintf("%s_%s.tar", fork.Parent.Name, strings.TrimPrefix(vol.Name, project.Name+"_"))
		_, err := os.Stat(filepath.Join(dir, expectedTarball))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		inputPath := filepath.Join("/cache", expectedTarball)
		outputPath := filepath.Join("/out", vol.Name)
		commands = append(commands, inputPath, outputPath)

		// set up binds so container can find the volumes
		binds = append(binds, fmt.Sprintf("%s:/out/%s", vol.Name, vol.Name))
	}

	// Nothing to do, no volumes
	if len(binds) <= 1 {
		return nil
	}

	id, err := a.createSystemContainer(client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: commands,
		},
		HostConfig: &container.HostConfig{
			Binds:      binds,
			AutoRemove: false,
			UsernsMode: "host", // Pass through uids
		},
	})
	if err != nil {
		return err
	}

	waitResult := a.Client().ContainerWait(context.Background(), id, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitResult.Error:
		return err
	case status := <-waitResult.Result:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	_, err = a.Client().ContainerRemove(context.Background(), id, client.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}
