package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Commands for the current worktree",
}

var worktreeUpCmd = &cobra.Command{
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

var worktreeDownCmd = &cobra.Command{
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

var worktreePsCmd = &cobra.Command{
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

var worktreeExecCmd = &cobra.Command{
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

var worktreeRestartCmd = &cobra.Command{
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

func init() {
	worktreeExecCmd.Flags().SetInterspersed(false)

	worktreeCmd.AddCommand(worktreeUpCmd)
	worktreeCmd.AddCommand(worktreeDownCmd)
	worktreeCmd.AddCommand(worktreePsCmd)
	worktreeCmd.AddCommand(worktreeExecCmd)
	worktreeCmd.AddCommand(worktreeRestartCmd)

	rootCmd.AddCommand(worktreeCmd)
}
