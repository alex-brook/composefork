package internal

import (
	"context"
	"fmt"
	"io"
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

	binds := []string{}
	for _, vol := range project.Volumes {
		binds = append(binds, fmt.Sprintf("%s:/in/%s:ro", vol.Name, vol.Name))
	}

	id, err := createSystemContainer(cli, client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: []string{"sleep", "infinity"},
		},
		HostConfig: &container.HostConfig{
			Binds: binds,
		},
	})
	if err != nil {
		return err
	}

	for _, vol := range project.Volumes {
		log.Println("Caching", vol.Name)
		result, err := cli.Client().CopyFromContainer(context.Background(), id, client.CopyFromContainerOptions{
			SourcePath: filepath.Join("/", "in", vol.Name),
		})
		if err != nil {
			return fmt.Errorf("error copying volume %s: %w", vol.Name, err)
		}

		out, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s.tar", vol.Name)))
		if err != nil {
			return fmt.Errorf("error copying volume %s: %w", vol.Name, err)
		}
		defer out.Close()
		io.Copy(out, result.Content)
	}

	_, err = cli.Client().ContainerRemove(context.Background(), id, client.ContainerRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("error copying volumes: %w", err)
	}

	log.Println("Finished caching")
	return nil
}
