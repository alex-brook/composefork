package cmd

import (
	"errors"
	"os"

	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"
)

var registeredCommands []func() *cobra.Command

func register(c func() *cobra.Command) { registeredCommands = append(registeredCommands, c) }

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "composefork",
		Short: "Clone a docker compose project",
		Long: `
  [FOR AGENTS]

  Composefork duplicates an existing docker compose project, changing details
  that let it run in parallel with the original, so each git worktree gets an
  isolated copy of the project's development environment.

  Start here. Run "composefork help <command>" for one command, and
  "composefork setup" to retrofit a project for composefork.

  When to use it
    - You are working in a git worktree, not the main checkout.
    - The project uses a Docker Compose-based devcontainer.
    - You need the container environment running.

    Do NOT run "docker compose up" in a worktree: it would use the parent
    project's name and collide with other worktrees. The main checkout is the
    human's own environment, managed with plain "docker compose".

  The volume cache
    A first "composefork up" can be slow while services install dependencies.
    "composefork cache" front-loads that once, so later forks start fast; see
    "composefork help cache". Never remove a healthcheck to make start-up
    faster: the healthcheck is the signal that the install has finished.

  Running commands
    Run every fork command from the root of your worktree: the fork name and
    ".env" are taken from the current directory, so a subdirectory names the
    project differently or fails to find a compose file that only ".env" points
    at. The global commands ("cache", "ls", "version", "setup", "help") run
    anywhere.

    A fork is named {original_project}-{worktree_dirname}.`,
	}
	for _, cmdFunc := range registeredCommands {
		rootCmd.AddCommand(cmdFunc())
	}
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err == nil {
		return
	} else if codeErr, ok := errors.AsType[cli.StatusError](err); ok {
		os.Exit(codeErr.StatusCode)
	} else {
		os.Exit(1)
	}
}
