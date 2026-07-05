package internal

import (
	"context"
	"fmt"
	// "io"
	"log"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func RunCacheCommand() error {
	docker, compose, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing docker client: %w", err)
	}

	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	err = compose.Down(context.Background(), name, api.DownOptions{
		Volumes:       false,
		RemoveOrphans: true,
	})
	if err != nil {
		return fmt.Errorf("error tearing down project: %w", err)
	}

	err = exportVolumes(docker, parent)
	if err != nil {
		return err
	}

	return nil
}

func exportVolumes(cli *command.DockerCli, project *types.Project) error {
	dir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("error copying volumes: %w", err)
	}
	dir = filepath.Join(dir, APP_NAME)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error copying volumes: %w", err)
	}

	binds := []string{fmt.Sprintf("%s:/out", dir)}
	commands := []string{"sh", "-c", `
    set -e
    echo "all args $@"
    while [ "$#" -ge 2 ]; do
      echo "copying $1 to $2"
      tar -c --numeric-owner -f $2 -C $1 .
      shift 2
    done
  `, "foo"}
	for _, vol := range project.Volumes {
		binds = append(binds, fmt.Sprintf("%s:/in/%s:ro", vol.Name, vol.Name))
		inputPath := filepath.Join("/in", vol.Name)
		outputPath := filepath.Join("/out", fmt.Sprintf("%s.tar", vol.Name))
		commands = append(commands, inputPath, outputPath)
	}

	id, err := createSystemContainer(cli, client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: commands,
		},
		HostConfig: &container.HostConfig{
			Binds:      binds,
			AutoRemove: true,
			UsernsMode: "host",
		},
	})
	if err != nil {
		return err
	}

	waitResult := cli.Client().ContainerWait(context.Background(), id, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitResult.Error:
		return err
	case status := <-waitResult.Result:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	log.Println("Finished caching")
	return nil
}
