package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List forked projects across all worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Ls(cmd.OutOrStdout())
		},
	}
	return lsCmd
}

func init() {
	register(newLsCmd)
}
