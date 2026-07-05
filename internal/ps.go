package internal

import (
	"fmt"
)

func (a *App) Ps() error {
	// Resolve parent / master project
	fork, err := LoadFork()
	if err != nil {
		return fmt.Errorf("error loading project: %w", err)
	}

	return a.printProjectStatus(fork.Name)
}
