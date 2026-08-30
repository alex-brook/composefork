# TODO

Open items ranked by severity, worst first. Completed work is at the bottom.

> 🤖 Ranking and the 🤖-marked notes below are Claude's; everything else is
> unchanged. Severity here means blast radius on the developer's real
> environment, then how badly a wrong result misleads an agent, then the rest.

## Critical — silent data loss

- [] Running `composefork up` in the main worktree should be equivalent to docker compose up
    - [] It isn't: `up` there restores the cache over the project's live volumes
        - `Up` calls `importVolumes` unconditionally, and in the main worktree the expected
          tarball name resolves to the parent's own, so `cache` followed by `up` silently
          replaces live data — losing local db state via postgres_data
        - `Project.Root()` is the guard, it just isn't applied
    - [] Same cause, wider blast radius: import runs on *every* `up`, so re-running
      `up` on a live fork restores the snapshot over that fork's current volumes too

- [] Volumes with an explicit `name:` or `external: true` are cached but never restored
    - They carry no project prefix, so the name we look for on import never matches
    - 🤖 Ranked critical, not medium: the missing restore is the harmless half. Because
      `applyForkOverrides` never rewrites them either, the fork *mounts the parent's
      actual volume* — it reads and writes the developer's live database, and whichever
      project created the volume first owns the label `down --volumes` selects on, so a
      fork's teardown can delete it outright

- [] `prune` is daemon wide
    - [] It removes forks belonging to other repositories
    - [] It deletes the main project's volumes if its directory was moved
    - [] Forks that were already `down`ed have no containers left, so it can't see their
      volumes and never cleans them

## High — the agent gets stuck or is actively misled

- [] Crashed services vanish from `ps` instead of showing as exited
    - `Ps` is called with `All: false`, so a dead container is simply absent
    - This is the "agent got stuck" case again — restart only helps if you know to run it

- [] `up` waits for health with no timeout, a never-healthy service hangs forever with no output

## Medium — correctness and ergonomics

- [] Setup prompt that covers:
    - [] That a compose project exists
    - [] That it can be recognised from the root directory (.env)
    - [] That healthchecks are defined for every service that installs deps on start, st when the health check passes the deps are ready to be cached
    - [] Adding a claude hook that runs `composefork worktree down` on SessionEnd if composefork is installed
    - [] Adding context to AGENTS.md or CLAUDE.md about composefork, and how to use `composefork skill` to get further context
    - 🤖 Worth pulling forward: the missing `.env` is not just a setup nicety. Without a
      pinned `COMPOSE_PROJECT_NAME` the parent name is inferred per worktree, which
      silently turns the entire volume cache into a no-op

- [] Commands have no `Args` validators, stray arguments are silently ignored

- [] Cached snapshots are never pruned, the cache dir grows unbounded

- [] Errors print the whole usage block, only `exec` sets `SilenceUsage`

## Low — cosmetics and docs

- [] Add a verbose flag

- [] Small cleanups
    - [] `internal/system_image_test.go` comment describes `/import` and `/export` links
      dispatching on argv[0]; it's really one `/runner` entrypoint dispatching on argv[1]
    - [] Cache snapshots are gzipped but named `.tar`
    - [] `dir` params in `exportVolumes` and `withDirLock` are shadowed by a fresh
      `cacheDir()` call, and the lock only covers the rename, not the export
    - [] `cmd/prune.go` short help: "Remove orphaned project that have had their worktree deleted"
    - [] `internal/app.go` comment mentions "interactive exec", stale since exec went
      non-interactive, and probably the origin of the wrong skill text
    - [] This file still says `composefork worktree <cmd>` in places, subcommands are flat now

## Future work

- Investigate checkpoints to avoid using too much RAM
    - [] Guess memory consumption based on avg. of existing projects
    - [] Decide if we will go over the allocated docker RAM with a new fork
    - [] Checkpoint/pause the oldest
    - [] When you interact with a paused project, prints a note for agent to ask permission before unpausing

## Done

- It wasn't obvious that `composefork worktree up` waits for the project to be healthy
    - [x] Show health as a column in project info

- [x] When a project container died, the agent got stuck
    - [x] Add `composefork worktree restart` command

- [x] The agent confused ls and ps, is there misleading docs?

- [x] There is no exec command, the agent has to construct a vanilla compose command with the project name

- [x] Add version command

- [x] Volumes are not copied on fork, which means they take a long time to start
    - [x] Add a new command `composefork cache`
    - [x] Stop the parent project
        - 🤖 Stale: `cache` builds a throwaway randomly-named project instead of stopping
          the parent. Good — but the teardown for it only runs on the success path, so
          every failed `cache` leaks a full duplicate stack that nothing can reclaim
    - [x] System container for these kinds of operations
    - [x] Snapshot volumes
    - [x] Use these cached volumes when creating forks

- [x] Each fork should build its own image, not just use the parent image
    - [x] Fork images should be removed when the project is torn down
        - 🤖 Only true for services that omit `image:`. With an explicit tag the fork
          builds over the shared tag and `down` then deletes it out from under the parent

- [x] Expose forked services on 127.0.0.1
    - 🤖 `applyForkOverrides` sets `HostIP` to `127.0.0.1` instead of clearing it.
      Clearing it wasn't enough: a bare `3000:3000` leaves the field empty and the daemon
      then binds `0.0.0.0`, so the fork has to force loopback rather than preserve what
      the parent authored
    - 🤖 Covered by `TestApplyForkOverridesBindsLoopback` (bare, explicit loopback,
      wildcard, LAN address, udp, expanded range) and by port assertions on the existing
      bring-ups in `TestForkUp` and `TestWorktree` — the latter guards that the main
      worktree keeps its own authored bindings

- [x] Replace bundled debian with smaller non-gpl image

- [x] Add tests to CI
