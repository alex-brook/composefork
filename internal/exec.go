package internal

import (
	"context"
	"fmt"

	"github.com/docker/compose/v5/pkg/api"
)

func RunExecCommand(service string, command []string) error {
	_, compose, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing docker client: %w", err)
	}

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

	_, err = compose.Exec(context.Background(), name, opts)
	if err != nil {
		return fmt.Errorf("error executing: %w", err)
	}

	return nil
}
