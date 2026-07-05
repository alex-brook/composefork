package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Restart(services []string) error {
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	err = a.Compose.Restart(context.Background(), name, api.RestartOptions{
		Services: services,
	})
	if err != nil {
		return fmt.Errorf("error restarting: %w", err)
	}

	return nil
}
