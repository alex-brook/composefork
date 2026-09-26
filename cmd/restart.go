package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart [service...]",
		Short: "Restart the compose project for this worktree",
		Long: `Restarts the running containers of the current worktree's forked compose project,
preserving the dynamically assigned host ports. Pass one or more service names to
restart only those services; with no arguments the whole project is restarted.

This restarts processes only: it does not pick up changes to the compose file or
the Dockerfile. Use "composefork down" then "composefork up" for those.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Restart(args)
		},
	}
	return restartCmd
}

func init() {
	register(newRestartCmd)
}
