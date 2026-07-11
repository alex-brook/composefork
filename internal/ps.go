package internal

import (
	"fmt"
)

func (a *App) Ps() error {
	// Resolve parent / master project
	project, err := NewProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	return a.printProjectStatus(project.Name)
}
