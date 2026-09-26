package cmd

import (
	"errors"

	"github.com/alex-brook/composefork/internal"
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	execCmd := &cobra.Command{
		Use:   "exec [service] [command] [args...]",
		Short: "Execute a command against a service",
		Long: `Runs a one-off command inside a running service container of the current
worktree's forked project, for example "composefork exec app ls /".

This is not a shell session: no TTY is attached and stdin is not connected, so
"composefork exec app bash" exits immediately and anything that waits for input
will hang or fail. Pass the command and its arguments directly instead.

Requires both a service name and a command, and the service must already be
running. Anything after the command, including flags, is passed straight through
to the command rather than interpreted by composefork.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			service := args[0]
			command := args[1:]

			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			err = app.Exec(service, command)
			if err == nil {
				return nil
			} else if statusErr, ok := errors.AsType[cli.StatusError](err); ok {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return statusErr
			}

			return err
		},
	}
	execCmd.Flags().SetInterspersed(false)
	return execCmd
}

func init() {
	register(newExecCmd)
}
