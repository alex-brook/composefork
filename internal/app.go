package internal

import (
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	"github.com/moby/moby/client"
)

type App struct {
	Docker  *command.DockerCli
	Compose api.Compose
}

func NewApp() (*App, error) {
	docker, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("error creating docker cli: %w", err)
	}
	if err := docker.Initialize(&flags.ClientOptions{}, command.WithOutputStream(os.Stdout)); err != nil {
		return nil, fmt.Errorf("error initializing docker cli: %w", err)
	}

	service, err := compose.NewComposeService(docker, compose.WithOutputStream(os.Stdout))
	if err != nil {
		return nil, fmt.Errorf("error creating compose service: %w", err)
	}

	return &App{Docker: docker, Compose: service}, nil
}

func (a *App) Client() client.APIClient {
	return a.Docker.Client()
}
