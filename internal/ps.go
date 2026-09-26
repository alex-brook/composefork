package internal

import (
	"fmt"
)

func (a *App) Ps() error {
	if err := requireWorktree(); err != nil {
		return err
	}

	// Resolve parent / master project
	project, err := NewProject("")
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	return a.printProjectStatus(project.Name)
}
