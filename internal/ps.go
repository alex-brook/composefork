package internal

import (
	"fmt"
)

func (a *App) Ps() error {
	// Resolve parent / master project
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	return a.printProjectStatus(name)
}
