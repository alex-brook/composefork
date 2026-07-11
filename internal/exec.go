package internal

import (
	"context"
	"fmt"
	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Exec(service string, command []string) error {
	project, err := NewProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	opts := api.RunOptions{
		Service:     service,
		Command:     command,
		Tty:         false,
		Interactive: false,
	}

	_, err = a.Compose.Exec(context.Background(), project.Name, opts)
	if err != nil {
		return err
	}

	return nil
}
