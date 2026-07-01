package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
)

const COMPOSEFORK_LABEL = "com.github.alex-brook.composefork"

func newClient() (*command.DockerCli, api.Compose, error) {
	client, err := command.NewDockerCli()
	if err != nil {
		return nil, nil, err
	}
	client.Initialize(&flags.ClientOptions{}, command.WithOutputStream(os.Stdout))

	service, err := compose.NewComposeService(client, compose.WithOutputStream(os.Stdout))
	if err != nil {
		return nil, nil, err
	}

	return client, service, err
}

func projectName(parent *types.Project) (string, string, error) {
	oldName := parent.Name

	wd, err := os.Getwd()
	if err != nil {
		return "", oldName, err
	}
	parts := strings.Split(wd, "/")
	dirname := parts[len(parts)-1]

	return fmt.Sprintf("%s_%s", oldName, dirname), oldName, nil
}

func loadProject() (*types.Project, error) {
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

	project, err := opts.LoadProject(context.Background())
	if err != nil {
		return nil, err
	}

	return project, nil
}

func printProjectStatus(compose api.Compose, name string) error {
	containers, err := compose.Ps(context.Background(), name, api.PsOptions{All: false})
	if err != nil {
		return fmt.Errorf("error in ps: %w", err)
	}

	for _, container := range containers {
		var ports strings.Builder
		published := 0
		for _, publisher := range container.Publishers {
			if publisher.PublishedPort == 0 {
				continue
			}
			if published >= 1 {
				ports.WriteString(", ")
			}
			portsLine := fmt.Sprintf("%s:%d->%d/%s", publisher.URL, publisher.PublishedPort, publisher.TargetPort, publisher.Protocol)
			ports.WriteString(portsLine)
			published += 1
		}

		fmt.Printf("%-10s %-12s %s\n", container.Service, container.State, ports.String())
	}

	return nil
}
