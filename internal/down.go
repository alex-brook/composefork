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
	project, err := NewProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	log.Println("Tearing down project", project.Name)
	err = a.Compose.Down(context.Background(), project.Name, api.DownOptions{
		Volumes:       !project.Root,
		RemoveOrphans: true,
		Images:        "local",
	})
	if err != nil {
		return fmt.Errorf("error tearing down project: %w", err)
	}

	return nil
}
