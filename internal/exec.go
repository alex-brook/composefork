package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Exec(service string, command []string) error {
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	opts := api.RunOptions{
		Service:     service,
		Command:     command,
		Tty:         true,
		Interactive: true,
	}

	_, err = a.Compose.Exec(context.Background(), name, opts)
	if err != nil {
		return fmt.Errorf("error executing: %w", err)
	}

	return nil
}
