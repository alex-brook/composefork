/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
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
  Composefork duplicates an existing docker compose project, changing details that
  allow it to run in parallel with the original. This is useful for running parallel
  agents with your existing devcontainer environment`,
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
