package cmd

import (
	"github.com/alex-brook/composefork/internal"
	"github.com/spf13/cobra"
)

func newUpCmd() *cobra.Command {
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Bring up the compose project for this worktree",
		Long: `Brings up a forked copy of the compose project for the current worktree.

Published ports are rebound to 127.0.0.1 on dynamically assigned host ports (no
fixed host ports), so multiple worktrees can run simultaneously without conflict
and are not reachable from the local network.

It waits until every service is healthy, or merely running for a service that
defines no healthcheck, then prints the same table as "composefork ps". Expect
the first run to take a while: services that build get their own per-fork image,
while a service pinned to an explicit "image:" tag shares that image with the
main checkout and other forks. If start-up stays slow, the project's volume
cache has not been built; a human can run "composefork cache" (see
"composefork help cache").

Files listed in the project's ".worktreeinclude" that git leaves out of the
worktree (gitignored files such as ".env" or a local database) are seeded from
the main checkout. Claude Code copies them when it creates the worktree;
"composefork up" fills in whatever it did not, without overwriting edits the
fork has made.

Run this before starting work that requires the container environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := internal.NewApp(cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return app.Up()
		},
	}
	return upCmd
}

func init() {
	register(newUpCmd)
}
