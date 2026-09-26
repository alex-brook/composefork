# TODO

Open items ranked by severity, worst first. Completed work is at the bottom.
Format: one short bullet per item; nested bullets only for supporting detail.

## Critical — hits a real developer, silently

- [ ] Volumes with explicit `name:` or `external: true` are cached but never restored
    - No project prefix, so the import name never matches
    - Worse: the fork mounts the parent's live volume, and `down --volumes` can delete it

## High — the agent gets stuck or is actively misled

- [ ] `up` waits for health with no timeout; a never-healthy service hangs forever with no output

## Medium — correctness and ergonomics

- [ ] Setup prompt should cover:
    - [ ] That a compose project exists
    - [ ] That it can be recognised from the root directory (`.env`)
    - [ ] That every service installing deps on start has a healthcheck; deps are cacheable once healthy
    - [ ] A Claude hook running `composefork worktree down` on SessionEnd when installed
    - [ ] AGENTS.md/CLAUDE.md context on composefork and `composefork skill`
    - Missing `.env` is not cosmetic: without a pinned `COMPOSE_PROJECT_NAME`, the parent name is inferred per worktree and the volume cache becomes a no-op

- [ ] Support bare repositories, where every checkout is a linked worktree and none is the main one
    - [ ] `cache` chdirs to `projectRoot()`, which returns a bare `/repo.git` unchanged — no working tree
    - [ ] Nothing is the parent, so there are no volumes to snapshot and no `.worktreeinclude` source

- [ ] Commands have no `Args` validators; stray arguments are silently ignored

- [ ] Cached snapshots are never pruned; the cache dir grows unbounded

- [ ] Errors print the whole usage block; only `exec` sets `SilenceUsage`

## Low — cosmetics and docs

- [ ] Add a verbose flag

- [ ] Small cleanups:
    - [ ] `internal/system_image_test.go` comment describes `/import`/`/export` by argv[0]; it is one `/runner` entrypoint dispatching on argv[1]
    - [ ] Cache snapshots are gzipped but named `.tar`
    - [ ] `dir` params in `exportVolumes` and `withDirLock` are shadowed by `cacheDir()`, and the lock only covers the rename, not the export
    - [ ] `internal/app.go` comment mentions "interactive exec", stale since exec went non-interactive (likely origin of the wrong skill text)
    - [ ] This file still says `composefork worktree <cmd>` in places; subcommands are flat now

## Future work

- [ ] Investigate checkpoints to avoid using too much RAM
    - [ ] Estimate memory use from the average of existing projects
    - [ ] Decide if a new fork exceeds the allocated Docker RAM
    - [ ] Checkpoint/pause the oldest
    - [ ] On interaction with a paused project, print a note asking the agent to get permission before unpausing

- [ ] Use composefork to bring up a development stack in CI
    - Reliability, not speed. Guarantee: unchanged config survives an outage
    - CI restores the cache dir; `composefork ci` imports it; `cache` re-produces only on compose-config or lockfile change
    - Three artifacts in one dir: volume snapshots, image tarballs, build cache
    - Consumer stays on the default docker driver; producer needs a `docker-container` builder to export cache
    - [ ] `composefork ci` — restore everything cached and bring up the main worktree
        - Thin wrapper over `up`'s path, not a parallel one
        - Natural home for the "report what was restored" line
        - Read-only; never produces
        - Name describes the caller, not the behaviour (a cold-start restore is not CI)
    - [ ] Images — pulled service images plus base images, `ImageLoad` before `Create`
        - Built images out of scope; the build replaces them
        - Base images need `FROM` parsing, or capture during `cache`
    - [ ] Build cache — `cache_to` in `cache`, `cache_from` in `up`; `type=local`, `mode=max`, one dir per service
    - [ ] Producer branches: images only if build cache is also exportable
        - [ ] `cache` makes its own `docker-container` builder first
        - [ ] Report what was emitted; every failure mode here is silent
    - [ ] Define cache dir addressability, size ceiling, and cache-miss behaviour

- [ ] Decide whether podman is supported, and make it true either way
    - [ ] Does it build? `BuildKitEnabled()` is true on podman, so compose takes the bake path against a daemon without BuildKit
    - [ ] Volume snapshot round trip under rootless podman (`UsernsMode: "host"` vs subuid mapping)
    - [ ] Then document as supported and add to CI, or detect and refuse clearly

## Done

- [x] `composefork up` in the main worktree should equal `docker compose up`
    - Decided instead to refuse: every fork command (`up`, `down`, `ps`, `exec`, `restart`) runs only in a worktree and tells the caller to create and enter one first; `cache`, `ls`, `version` and `skill` stay global
    - Resolves the fresh-clone snapshot import, and lets `Project.Root()` and its branches go entirely

- [x] Untracked files listed in `.worktreeinclude` are missing when an agent makes the worktree
    - [x] Copy the missing ones from the main worktree on `up`, before the compose project is loaded
    - git does the matching (no new dependency); seeds only absent files; per-entry failures collected as warnings
    - Covered by `internal/worktreeinclude_test.go`, `TestForkUpWithoutIncludedFiles`, `TestForkUpTwiceKeepsLocalEdits`

- [x] It wasn't obvious that `composefork worktree up` waits for the project to be healthy
    - [x] Show health as a column in project info

- [x] When a project container died, the agent got stuck
    - [x] Add `composefork worktree restart`

- [x] The agent confused `ls` and `ps`; is there misleading docs?

- [x] There is no exec command; the agent has to construct a vanilla compose command with the project name

- [x] Add version command

- [x] Volumes are not copied on fork, so they take a long time to start
    - [x] Add `composefork cache`
    - [x] Stop the parent project (now builds a throwaway randomly-named project instead)
    - [x] System container for these kinds of operations
    - [x] Snapshot volumes
    - [x] Use cached volumes when creating forks

- [x] Each fork should build its own image, not just use the parent image
    - [x] Fork images should be removed when the project is torn down
        - Only true for services that omit `image:`; an explicit tag is shared and `down` deletes it out from under the parent

- [x] Expose forked services on 127.0.0.1
    - `applyForkOverrides` sets `HostIP` to `127.0.0.1` rather than clearing it, so a bare `3000:3000` cannot bind `0.0.0.0`
    - Covered by `TestApplyForkOverridesBindsLoopback`, plus port assertions in `TestForkUp`/`TestWorktree`

- [x] Crashed services vanish from `ps` instead of showing as exited
    - `Ps` used `All: false`; `printProjectStatus` now passes `All: true`
    - Covered by `TestPsShowsCrashedService`

- [x] Replace bundled Debian with a smaller non-GPL image

- [x] Add tests to CI
