package internal

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/client"
)

func (a *App) Prune() error {
	cl := a.Client()

	f := client.Filters{}
	f.Add("label", COMPOSEFORK_PROJECT_LABEL)

	containers, err := cl.ContainerList(context.Background(), client.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return fmt.Errorf("error listing containers: %w", err)
	}

	set := make(map[string]string)
	for _, item := range containers.Items {
		projectName := item.Labels[api.ProjectLabel]
		_, ok := set[projectName]
		if ok {
			continue
		}
		location := item.Labels[COMPOSEFORK_DIR_LABEL]
		set[projectName] = location
	}

	for key, val := range set {
		_, err := os.Stat(val)
		if err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		fmt.Println("Removing ", key)
		err = a.Compose.Down(context.Background(), key, api.DownOptions{
			Volumes:       true,
			RemoveOrphans: true,
			Images:        "local",
		})
		if err != nil {
			return err
		}

	}

	return nil
}
