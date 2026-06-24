package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Fork the project for this worktree",
	Long: `
    Assuming you are in a worktree of a project that defines a compose based
    devcontainer, bring up a fork of that project for use in the worktree
  `,
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

var worktreePsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List compose projects for all worktrees",
}

var worktreePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove all compose projects for worktrees that no longer exist",
}

func init() {
	worktreeCmd.AddCommand(worktreeUpCmd)
	worktreeCmd.AddCommand(worktreeDownCmd)
	worktreeCmd.AddCommand(worktreePsCmd)
	worktreeCmd.AddCommand(worktreePruneCmd)

	rootCmd.AddCommand(worktreeCmd)
}
