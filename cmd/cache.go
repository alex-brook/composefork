package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache the volumes of your main project to improve fork start up time",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := internal.NewApp()
		if err != nil {
			return err
		}
		return app.Cache()
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
