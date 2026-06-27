package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Commands for the current worktree",
}

var worktreeUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring up the compose project for this worktree",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunUpCommand()
	},
}

var worktreeDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the compose project for this worktree",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunDownCommand()
	},
}

func init() {
	worktreeCmd.AddCommand(worktreeUpCmd)
	worktreeCmd.AddCommand(worktreeDownCmd)

	rootCmd.AddCommand(worktreeCmd)
}
