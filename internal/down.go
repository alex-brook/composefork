package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/compose/v5/pkg/api"
)

func RunDownCommand() error {
	// Initialize docker client
	_, compose, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing docker client: %w", err)
	}

	// Resolve parent / master project
	log.Println("Loading parent project")
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	log.Println("Tearing down project", name)
	err = compose.Down(context.Background(), name, api.DownOptions{
		Volumes:       true,
		RemoveOrphans: true,
		Images:        "local",
	})
	if err != nil {
		return fmt.Errorf("error tearing down project: %w", err)
	}

	return nil
}
