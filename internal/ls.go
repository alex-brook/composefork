package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/client"
)

func RunLsCommand() error {
	docker, _, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing the docker client: %w", err)
	}
	cl := docker.Client()

	f := client.Filters{}
	f.Add("label", COMPOSEFORK_LABEL)

	containers, err := cl.ContainerList(context.Background(), client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	set := make(map[string]string)
	results := []string{}
	for _, item := range containers.Items {
		projectName := item.Labels[api.ProjectLabel]
		_, ok := set[projectName]
		if ok {
			continue
		}

		set[projectName] = ""
		results = append(results, projectName)
	}

	for _, result := range results {
		fmt.Println(result)
	}

	return nil
}
