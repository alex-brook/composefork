package internal

import (
	"fmt"
	"io"
	"log"
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
	Out     io.Writer
	Log     *log.Logger
}

func NewApp(out io.Writer) (*App, error) {
	docker, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("error creating docker cli: %w", err)
	}
	if err := docker.Initialize(&flags.ClientOptions{}, command.WithOutputStream(out), command.WithErrorStream(out)); err != nil {
		return nil, fmt.Errorf("error initializing docker cli: %w", err)
	}

	// out+err both go to out so exec/build stderr is captured too; keep os.Stdin
	// for interactive exec. Setting compose streams is what makes exec honor out —
	// the lib only wraps the CLI's streams when they're provided here.
	service, err := compose.NewComposeService(docker, compose.WithStreams(out, out, os.Stdin))
	if err != nil {
		return nil, fmt.Errorf("error creating compose service: %w", err)
	}

	return &App{Docker: docker, Compose: service, Out: out, Log: log.New(out, "", log.LstdFlags)}, nil
}

func (a *App) Client() client.APIClient {
	return a.Docker.Client()
}
