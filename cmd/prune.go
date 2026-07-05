package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove orphaned project that have had their worktree deleted",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := internal.NewApp()
		if err != nil {
			return err
		}
		return app.Prune()
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
