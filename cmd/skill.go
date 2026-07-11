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

All commands must be run from inside your worktree directory.

The forked project is named {original_project}_{worktree_dirname}, so each
worktree gets its own isolated namespace automatically.

### composefork up
Brings up a forked copy of the compose project for the current worktree.
Port bindings are assigned dynamically (no fixed host ports), so multiple
worktrees can run simultaneously without conflict.

Run this before starting work that requires the container environment.

### composefork down
Tears down the forked compose project for the current worktree, including
volumes. Run this when you are done or need a clean environment.

### composefork ps
Lists the containers for the current worktree's forked project, showing each
service's state and the dynamically assigned host ports. Because ports are
assigned dynamically at "up" time, use this to find which host ports your
services are reachable on.

### composefork exec [service] [command...]
Runs a command inside a running service container of the current worktree's
forked project. The session is interactive with a TTY attached, so you can open
a shell (e.g. "composefork exec app bash") or run a one-off command.
Anything after the command, including flags, is passed straight through to the
command rather than interpreted by composefork.

### composefork restart [service...]
Restarts the running containers of the current worktree's forked compose
project, preserving the dynamically assigned host ports. Pass one or more
service names to restart only those services; with no arguments the whole
project is restarted. Use this to pick up changes that require a container
restart without a full down/up cycle.

### composefork ls
Lists all forked compose projects across all worktrees.

### composefork prune
Removes forked compose projects whose worktree directories no longer exist.
Run this to clean up after deleting a worktree.
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
