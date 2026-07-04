package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List forked projects across all worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunLsCommand()
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
