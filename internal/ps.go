package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/client"
)

func RunPsCommand() error {
	docker, _, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing the docker client: %w", err)
	}

	f := client.Filters{}
	f.Add("label", COMPOSEFORK_LABEL)

	result, err := docker.Client().ContainerList(context.Background(), client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	set := make(map[string]string)
	for _, item := range result.Items {
		projectName := item.Labels[api.ProjectLabel]
		set[projectName] = ""
	}

	for k := range set {
		fmt.Println(k)
	}

	return nil
}
