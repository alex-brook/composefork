package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "List this worktree's services and their ports",
		Long: `Lists the services for the current worktree's project, showing each service's
state, health and dynamically assigned host ports. Because ports are assigned
dynamically at "up" time, use this to find which host ports your services are
reachable on.

Exited services stay listed, so a crashed service shows as an exited row rather
than silently dropping out of the table.`,
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
