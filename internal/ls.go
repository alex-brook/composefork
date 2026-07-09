package internal

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/client"
)

func (a *App) Ls(w io.Writer) error {
	cl := a.Client()

	f := client.Filters{}
	f.Add("label", COMPOSEFORK_PROJECT_LABEL)

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
		fmt.Fprintln(w, result)
	}

	return nil
}
