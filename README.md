# Composefork

`composefork` is a small utility that lets you use your existing docker compose
based devcontainer configuration with parallel agents. It works by cloning the
original compose project and namespacing it for a single worktree

## Usage

### Start a forked environment for a worktree

```sh
git worktree add ../my-feature my-feature-branch
cd ../my-feature
composefork up
```

This creates another compose project named `<original>_my-feature` with
dynamically assigned ports, so it won't conflict with other worktrees.

### Tear it down when you're done

```sh
composefork down
```

### See what's running (and on which ports)

```sh
composefork ps
```

Lists the services for the current worktree's project with their state and the
dynamically assigned host ports.

### Run a command in a service

```sh
composefork exec <service> <command> [args...]
```

Runs a one-off command inside a running service container for the current
worktree — for example `composefork exec app ls /`. The command runs
non-interactively (no TTY); anything after the command, including flags, is
passed straight through rather than interpreted by composefork.

### Restart the running environment

```sh
composefork restart [service...]
```

Restarts the containers for the current worktree's project, keeping the same
dynamically assigned ports. Pass service names to restart only those services,
or none to restart everything.

### Speed up fork start-up by caching volumes

```sh
composefork cache
```

Builds the main project once and snapshots its volumes to a local cache. Later
`composefork up` runs seed their volumes from that snapshot, so forks that would
otherwise reinstall dependencies on start come up fast. Run it from the main
worktree after dependencies change.

### See all forked projects

```sh
composefork ls
```

### Clean up after deleting a worktree

```sh
composefork prune
```

Removes any forked compose projects whose worktree directories no longer exist.

## Other commands

- `composefork version` — print version, commit, and build date.
- `composefork skill` — print agent-oriented usage instructions (intended to be
  fed to coding agents so they use composefork instead of raw `docker compose`).
