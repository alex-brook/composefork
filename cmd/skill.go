package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const skillText = `# composefork

## Why this tool exists

When multiple agents work in parallel on the same repository, each in its own
git worktree, they share a single Docker Compose project. Running that project
normally causes port conflicts (two services trying to bind the same host port)
and volume collisions (containers stomping on each other's data).

composefork solves this by forking the compose project for your worktree: it
creates a namespaced copy with a unique project name and dynamically assigned
ports, so each worktree runs a fully isolated container environment.

## When to use it

Use composefork when:
- You are working in a git worktree (not the main checkout)
- The project uses a Docker Compose-based devcontainer
- You need the container environment running to do your work

Do NOT run "docker compose up" directly — that would conflict with other
worktrees sharing the same project.

## Commands

Run every command from the root of your worktree. Running one from a
subdirectory silently creates a separate, wrong project.

The forked project is named {original_project}-{worktree_dirname}, so each
worktree gets its own isolated namespace automatically.

### composefork up
Brings up a forked copy of the compose project for the current worktree.
Port bindings are assigned dynamically (no fixed host ports), so multiple
worktrees can run simultaneously without conflict.

It blocks until every service is healthy, then prints the same table as
"composefork ps". Expect the first run to take a while, as your fork builds its
own images.

Run this before starting work that requires the container environment.

### composefork down
Tears down the forked compose project for the current worktree, removing its
containers, images and volumes. Run this when you are done or need a clean
environment.

In the main checkout there is no fork to remove, so it stops the project but
leaves your volumes alone.

### composefork ps
Lists the services for the current worktree's project, showing each service's
state, health and dynamically assigned host ports. Because ports are assigned
dynamically at "up" time, use this to find which host ports your services are
reachable on.

Only running services are listed. If a service you expect is missing, it has
crashed rather than been left out.

### composefork exec [service] [command...]
Runs a one-off command inside a running service container of the current
worktree's forked project, for example "composefork exec app ls /".

This is not a shell session: no TTY is attached and stdin is not connected, so
"composefork exec app bash" exits immediately and anything that waits for input
will hang or fail. Pass the command and its arguments directly instead.

Requires both a service name and a command, and the service must already be
running. Anything after the command, including flags, is passed straight
through to the command rather than interpreted by composefork.

### composefork restart [service...]
Restarts the running containers of the current worktree's forked compose
project, preserving the dynamically assigned host ports. Pass one or more
service names to restart only those services; with no arguments the whole
project is restarted.

This restarts processes only — it does not pick up changes to the compose file
or the Dockerfile. Use "composefork down" then "composefork up" for those.

### composefork cache
Saves a snapshot of the project's dependencies so that new forks start fast
instead of reinstalling them on every "up".

This is a setup command a human runs once for the project, not something you
need per worktree. It applies to the project as a whole, leaves your running
containers and local data untouched, and only helps when services define health
checks, since that is the signal that installing dependencies has finished.

"composefork up" works without a snapshot; it just starts slower.

### composefork ls
Lists all forked compose projects on this machine, across every repository.
Unlike "composefork ps", it is not scoped to the current worktree or project.

### composefork prune
Removes forked compose projects whose worktree directories no longer exist.
Run this to clean up after deleting a worktree.

### composefork version
Prints the composefork version, commit and build date.

### composefork skill
Prints this document.
`

func newSkillCmd() *cobra.Command {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "Agents: read this before using composefork",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), skillText)
		},
	}
	return skillCmd
}

func init() {
	register(newSkillCmd)
}
