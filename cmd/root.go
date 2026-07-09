/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var registeredCommands []func() *cobra.Command

func register(c func() *cobra.Command) { registeredCommands = append(registeredCommands, c) }

func newRootCmd() *cobra.Command {
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

type CodeError struct {
	Code int
	Err  error
}

func (e *CodeError) Error() string { return e.Err.Error() }

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := newRootCmd().Execute()
	var codeErr *CodeError
	if err == nil {
		return
	} else if errors.As(err, &codeErr) {
		os.Exit(codeErr.Code)
	} else {
		os.Exit(1)
	}
}
