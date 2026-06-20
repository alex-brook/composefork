package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
)

type App struct {
	docker  *command.DockerCli
	compose api.Compose
	opts    *cli.ProjectOptions
	project *types.Project
}

func NewApp() *App {
	client, err := command.NewDockerCli()
	if err != nil {
		panic(err)
	}
	client.Initialize(&flags.ClientOptions{}, command.WithOutputStream(os.Stdout))

	service, err := compose.NewComposeService(client, compose.WithOutputStream(os.Stdout))
	if err != nil {
		panic(err)
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
		panic(err)
	}

	return &App{docker: client, compose: service, opts: opts}
}

func main() {
	app := NewApp()

	// Resolve parent / master project
	log.Println("Loading parent project")
	parent, err := app.opts.LoadProject(context.Background())
	if err != nil {
		panic(err)
	}

	// Build the master project
	log.Println("Building", parent.Name)
	err = app.compose.Build(context.Background(), parent, api.BuildOptions{})
	if err != nil {
		panic(err)
	}

	// Create a child project for the worktree
	log.Println("Creating fork project")
	err = app.makeProject(parent)
	if err != nil {
		panic(err)
	}

	// Bring a child project up
	log.Println("Bringing up", app.project.Name)
	err = app.compose.Up(context.Background(), app.project, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{}})
	if err != nil {
		log.Fatalf("up error: %v", err)
	}
}

func (a *App) makeProject(parent *types.Project) error {
	oldName := parent.Name

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parts := strings.Split(wd, "/")
	dirname := parts[len(parts)-1]

	a.opts.Environment["COMPOSE_PROJECT_NAME"] = fmt.Sprintf("%s_%s", oldName, dirname)
	project, err := a.opts.LoadProject(context.Background())
	if err != nil {
		return err
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
		srv.CustomLabels[api.ProjectLabel] = project.Name
		srv.CustomLabels[api.ServiceLabel] = srv.Name
		srv.CustomLabels[api.OneoffLabel] = "False"
		srv.CustomLabels[api.WorkingDirLabel] = project.WorkingDir
		srv.CustomLabels[api.ConfigFilesLabel] = strings.Join(project.ComposeFiles, ",")
		project.Services[srv.Name] = srv
	}
	a.project = project

	return nil
}
