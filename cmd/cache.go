package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache the volumes of your main project to improve fork start up time",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Cache()
		},
	}
	return cacheCmd
}

func init() {
	register(newCacheCmd)
}
