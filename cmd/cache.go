package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache the volumes of your main project to improve fork start up time",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunCacheCommand()
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
