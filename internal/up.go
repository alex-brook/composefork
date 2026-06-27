package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
)

func RunUpCommand() error {
	// Initialize docker client
	_, compose, err := newClient()
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

		srv.CustomLabels[COMPOSEFORK_LABEL] = "1"

		project.Services[srv.Name] = srv
	}

	return project, nil
}
