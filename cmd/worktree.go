package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newWorktreeUpCmd() *cobra.Command {
	worktreeUpCmd := &cobra.Command{
		Use:   "up",
		Short: "Bring up the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Up()
		},
	}
	return worktreeUpCmd
}

func newWorktreeDownCmd() *cobra.Command {
	worktreeDownCmd := &cobra.Command{
		Use:   "down",
		Short: "Tear down the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Down()
		},
	}
	return worktreeDownCmd
}

func newWorktreePsCmd() *cobra.Command {
	worktreePsCmd := &cobra.Command{
		Use:   "ps",
		Short: "List this worktree's services and their ports",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Ps()
		},
	}
	return worktreePsCmd
}

func newWorktreeExecCmd() *cobra.Command {
	worktreeExecCmd := &cobra.Command{
		Use:   "exec [service] [command] [args...]",
		Short: "Execute a command against a service",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			service := args[0]
			command := args[1:]

			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Exec(service, command)
		},
	}
	worktreeExecCmd.Flags().SetInterspersed(false)
	return worktreeExecCmd
}

func newWorktreeRestartCmd() *cobra.Command {
	worktreeRestartCmd := &cobra.Command{
		Use:   "restart [service...]",
		Short: "Restart the compose project for this worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp()
			if err != nil {
				return err
			}
			return app.Restart(args)
		},
	}
	return worktreeRestartCmd
}

func newWorktreeCmd() *cobra.Command {
	worktreeCmd := &cobra.Command{
		Use:   "worktree",
		Short: "Commands for the current worktree",
	}

	worktreeCmd.AddCommand(newWorktreeUpCmd())
	worktreeCmd.AddCommand(newWorktreeDownCmd())
	worktreeCmd.AddCommand(newWorktreePsCmd())
	worktreeCmd.AddCommand(newWorktreeExecCmd())
	worktreeCmd.AddCommand(newWorktreeRestartCmd())

	return worktreeCmd
}

func init() {
	register(newWorktreeCmd)
}
