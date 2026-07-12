package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "List this worktree's services and their ports",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Ps()
		},
	}
	return psCmd
}

func init() {
	register(newPsCmd)
}
