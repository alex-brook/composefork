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
	project *types.Project
}

func NewApp() *App {
	client, err := command.NewDockerCli()
	if err != nil {
		panic(err)
	}
	client.Initialize(&flags.ClientOptions{})

	service, err := compose.NewComposeService(client)
	if err != nil {
		panic(err)
	}

	return &App{docker: client, compose: service}
}

func main() {
	app := NewApp()

	// Parse the original compose definition
	log.Println("Making project")
	err := app.makeProject()
	if err != nil {
		panic(err)
	}

	log.Println("Bringing up project")
	err = app.compose.Up(context.Background(), app.project, api.UpOptions{Create: api.CreateOptions{}, Start: api.StartOptions{}})
	if err != nil {
		log.Fatalf("up error: %v", err)
	}
}

func (a *App) makeProject() error {
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
		return err
	}

	discovered, err := opts.LoadProject(context.Background())
	if err != nil {
		return err
	}
	oldName := discovered.Name

	opts.Environment["COMPOSE_PROJECT_NAME"] = os.Args[1]
	project, err := opts.LoadProject(context.Background())
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
