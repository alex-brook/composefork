package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List forked projects across all worktrees",
		Long: `Lists all forked compose projects on this machine, across every repository.
Unlike "composefork ps", it is not scoped to the current worktree or project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Ls()
		},
	}
	return lsCmd
}

func init() {
	register(newLsCmd)
}
