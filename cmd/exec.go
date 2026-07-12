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
		Args:  cobra.MinimumNArgs(2),
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
