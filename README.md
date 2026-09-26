# Composefork

`composefork` is a small utility that lets you use your existing docker compose
based devcontainer configuration with parallel agents. It works by cloning the
original compose project and namespacing it for a single worktree

`composefork` complements a traditional development workflow. The original
worktree and compose project is reserved for you to work on manually.

## Prerequisites
- Docker installed on your system
- Git
- Claude Desktop or a similar tool that manages worktrees and chat threads
- A project configured with a docker compose devcontainer, check out the dummy app in `test/dummy` if you are unsure as to what this looks like — `test/dummy/.devcontainer/compose.yml` is the part worth copying.

## Installation
### Mise
```
mise use -g github:alex-brook/composefork@latest
```
### Manual
Download a binary matching your system from the releases page and add it to your PATH

## Getting started
- A static binary is available on the releases page
- Existing project? Run `composefork setup` and paste its output into a coding
  agent, or pipe it straight to a fresh Claude Code session with
  `composefork setup | claude -p`. The prompt walks the agent through the compose
  config, `.env`, `.worktreeinclude` and agent instructions, and ends with the
  assumptions composefork makes about your project
- Instruct agents to use `composefork` to manage the project in your AGENTS.md or similar. You may want to use `composefork skill` as a baseline
- Add a hook to your agent orchestrator to run `composefork down` when you close a chat thread
- Run `composefork cache` if your services have health checks, this makes bringing up a new stack faster and more reliable

## Reference

All commands below operate on a fork and only run inside a worktree; in the
main checkout they refuse and tell you to create and enter one first. `cache`
and `ls` are project- and machine-wide, and `version`/`skill` run anywhere.

### Start a forked environment for a worktree

```sh
cd ./.claude/worktrees/my-feature
composefork up
```

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

This restarts processes only — it does not pick up changes to the compose file
or the Dockerfile. Use `composefork down` then `composefork up` for those.

### Speed up fork start-up by caching volumes

```sh
composefork cache
```

Saves a snapshot of the project's installed dependencies to a local cache.
Later `composefork up` runs seed their volumes from that snapshot, so forks that
would otherwise reinstall dependencies on start come up fast.

It works on the project as a whole rather than on your current worktree, and it
leaves your running containers and their data untouched. Run it after
dependencies change.

### See all forked projects

```sh
composefork ls
```

## Other commands

- `composefork version` — print version, commit, and build date.
- `composefork skill` — print agent-oriented usage instructions (intended to be
  fed to coding agents so they use composefork instead of raw `docker compose`).
- `composefork setup` — print a prompt to paste into a coding agent to retrofit
  an existing project for composefork.
