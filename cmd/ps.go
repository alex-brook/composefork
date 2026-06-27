package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "View all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunPsCommand()
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
