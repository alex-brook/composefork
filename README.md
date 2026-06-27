# Composefork

`composefork` is a small utility that lets you use your existing docker compose
based devcontainer configuration with parallel agents. It works by cloning the
original compose project and namespacing it for a single worktree

## Usage

### Start a forked environment for a worktree

```sh
git worktree add ../my-feature my-feature-branch
cd ../my-feature
composefork worktree up
```

This creates an isolated compose project named `<original>_my-feature` with
dynamically assigned ports, so it won't conflict with other worktrees.

### Tear it down when you're done

```sh
composefork worktree down
```

### See all running forked projects

```sh
composefork ps
```

### Clean up after deleting a worktree

```sh
composefork prune
```

Removes any forked compose projects whose worktree directories no longer exist.
