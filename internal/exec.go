package internal

import (
	"context"
	"fmt"
	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Exec(service string, command []string) (int, error) {
	fork, err := LoadFork()
	if err != nil {
		return 1, fmt.Errorf("error loading project: %w", err)
	}

	opts := api.RunOptions{
		Service:     service,
		Command:     command,
		Tty:         false,
		Interactive: false,
	}

	code, err := a.Compose.Exec(context.Background(), fork.Name, opts)
	return code, err
}
