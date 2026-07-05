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

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func RunUpCommand() error {
	// Initialize docker client
	docker, compose, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing docker client: %w", err)
	}

	// Resolve parent / master project
	log.Println("Loading parent project")
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	// Build the master project
	log.Println("Building", parent.Name)
	err = compose.Build(context.Background(), parent, api.BuildOptions{})
	if err != nil {
		return fmt.Errorf("error building project: %w", err)
	}

	// Create a child project for the worktree
	log.Println("Creating fork project")
	project, err := overrideProject(parent)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	err = compose.Create(context.Background(), project, api.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	err = importVolumes(docker, parent, project)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	// Bring a child project up
	log.Println("Bringing up project", project.Name)
	err = compose.Up(context.Background(), project, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{Wait: true}})
	if err != nil {
		return fmt.Errorf("up error: %w", err)
	}

	// Print project info
	return printProjectStatus(compose, project.Name)
}

func overrideProject(parent *types.Project) (*types.Project, error) {
	name, oldName, err := projectName(parent)
	if err != nil {
		return nil, fmt.Errorf("error computing project name: %w", err)
	}

	opts, err := cli.NewProjectOptions(
		nil,
		cli.WithOsEnv,
		cli.WithEnvFiles(),
		cli.WithDotEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath,
		cli.WithResolvedPaths(true),
	)
	if err != nil {
		return nil, err
	}

	opts.Environment["COMPOSE_PROJECT_NAME"] = name
	project, err := opts.LoadProject(context.Background())
	if err != nil {
		return nil, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, srv := range project.Services {
		if srv.Image == "" {
			srv.Image = fmt.Sprintf("%s-%s:%s", oldName, srv.Name, "latest")
			srv.Build = &types.BuildConfig{}
		}
		for i := range srv.Ports {
			srv.Ports[i].Published = ""
			srv.Ports[i].HostIP = ""
		}
		if srv.CustomLabels == nil {
			srv.CustomLabels = types.Labels{}
		}
		// Set expected compose labels for this container
		srv.CustomLabels[api.ProjectLabel] = project.Name
		srv.CustomLabels[api.ServiceLabel] = srv.Name
		srv.CustomLabels[api.OneoffLabel] = "False"
		srv.CustomLabels[api.WorkingDirLabel] = project.WorkingDir
		srv.CustomLabels[api.ConfigFilesLabel] = strings.Join(project.ComposeFiles, ",")

		srv.CustomLabels[COMPOSEFORK_LABEL] = wd

		project.Services[srv.Name] = srv
	}

	return project, nil
}

func importVolumes(docker *command.DockerCli, parent *types.Project, project *types.Project) error {
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
		// set up binds so container can find the volumes
		binds = append(binds, fmt.Sprintf("%s:/out/%s", vol.Name, vol.Name))

		// set up the commands the container will run to copy the cached
		// snapshots into the child volumes
		expectedTarball := fmt.Sprintf("%s_%s.tar", parent.Name, strings.TrimPrefix(vol.Name, project.Name+"_"))
		fmt.Println(expectedTarball)
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
	}

	id, err := createSystemContainer(docker, client.ContainerCreateOptions{
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

	waitResult := docker.Client().ContainerWait(context.Background(), id, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitResult.Error:
		return err
	case status := <-waitResult.Result:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	_, err = docker.Client().ContainerRemove(context.Background(), id, client.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	fmt.Println(parent.Name, project.Name)

	return nil
}
