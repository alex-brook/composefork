package internal

import (
	"fmt"
)

func RunPsCommand() error {
	_, compose, err := newClient()
	if err != nil {
		return fmt.Errorf("error initializing the docker client: %w", err)
	}

	// Resolve parent / master project
	parent, err := loadProject()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	name, _, err := projectName(parent)
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	return printProjectStatus(compose, name)
}
