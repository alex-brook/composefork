package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed setup.md
var setupPrompt string

func newSetupCmd() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Print a prompt to set up a project for composefork",
		Long: `Prints a prompt for retrofitting an existing project to work with composefork.
Paste its output into a coding agent working in the project. A human runs this
once per project; it is not part of the per-worktree workflow.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), setupPrompt)
		},
	}
	return setupCmd
}

func init() {
	register(newSetupCmd)
}
