package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/compose/v5/pkg/api"
)

func (a *App) Down() error {
	// Resolve parent / master project
	log.Println("Loading parent project")
	fork, err := LoadFork()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	log.Println("Tearing down project", fork.Name)
	err = a.Compose.Down(context.Background(), fork.Name, api.DownOptions{
		Volumes:       true,
		RemoveOrphans: true,
		Images:        "local",
	})
	if err != nil {
		return fmt.Errorf("error tearing down project: %w", err)
	}

	return nil
}
