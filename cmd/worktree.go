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
	Run: func(cmd *cobra.Command, args []string) {
		internal.NewForkCommand().Run()
	},
}

func init() {
	rootCmd.AddCommand(worktreeCmd)
}
