package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
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

	// Bring a child project up
	log.Println("Bringing up", project.Name)
	err = compose.Up(context.Background(), project, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{}})
	if err != nil {
		return fmt.Errorf("up error: %w", err)
	}

	// Print project info
	cl := docker.Client()
	f := client.Filters{}
	f.Add("label", fmt.Sprintf("%s=%s", api.ProjectLabel, project.Name))

	containers, err := cl.ContainerList(context.Background(), client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	fmt.Println(project.Name, "ports")
	for _, cont := range containers.Items {
		exposedPorts := []container.PortSummary{}
		for _, port := range cont.Ports {
			if port.PublicPort == 0 {
				continue
			}
			exposedPorts = append(exposedPorts, port)
		}

		if len(exposedPorts) == 0 {
			continue
		}

		serviceName := cont.Labels[api.ServiceLabel]
		fmt.Printf("%2s%s\n", "", serviceName)
		for _, port := range exposedPorts {
			fmt.Printf("%4s%d:%d\n", "", port.PublicPort, port.PrivatePort)
		}
	}

	return nil
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
