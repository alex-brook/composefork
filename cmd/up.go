package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newUpCmd() *cobra.Command {
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Bring up the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Up()
		},
	}
	return upCmd
}

func init() {
	register(newUpCmd)
}
