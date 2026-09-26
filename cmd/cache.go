package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache the volumes of your main project to improve fork start up time",
		Long: `Builds the volume cache that lets new forks start fast instead of reinstalling
dependencies on every "up".

A first start is expensive because services install dependencies before they can
serve anything, and the compose file's healthcheck is what marks that install as
finished. The cost is the install, not the healthcheck.

This command pays that cost once and shares it. It starts a throwaway copy of the
project under a random name, waits for the healthchecks to go green, snapshots
its named volumes (which by then hold the installed dependencies, plus data
volumes such as the database in their freshly initialised state rather than the
developer's current data), then tears the copy down. Your own running containers
and data are left untouched.

A human runs this once per project and again whenever dependencies change; it is
not something you need per worktree. "composefork up" works without it, just
slower, and a healthy stack is not proof a cache exists.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Cache()
		},
	}
	return cacheCmd
}

func init() {
	register(newCacheCmd)
}
