package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Restart(services []string) error {
	project, err := NewProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	err = a.Compose.Restart(context.Background(), project.Name, api.RestartOptions{
		Services: services,
	})
	if err != nil {
		return fmt.Errorf("error restarting: %w", err)
	}

	return nil
}
