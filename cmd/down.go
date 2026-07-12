package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newDownCmd() *cobra.Command {
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Tear down the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Down()
		},
	}
	return downCmd
}

func init() {
	register(newDownCmd)
}
