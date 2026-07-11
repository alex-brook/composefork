package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart [service...]",
		Short: "Restart the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Restart(args)
		},
	}
	return restartCmd
}

func init() {
	register(newRestartCmd)
}
