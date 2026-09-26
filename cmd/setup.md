# composefork setup prompt

You are setting up this repository so it works with `composefork`, a tool that
gives every git worktree its own isolated copy of the project's docker compose
environment. When you are done, a developer should be able to create a git
worktree, run `composefork up` inside it, and get a working stack that does not
collide with the main checkout or with any other worktree.

Work through the steps in order. Inspect the repository before editing, keep
changes minimal, preserve existing behaviour, and ask before overwriting files or
committing anything. Never commit secrets. If a step does not apply, say so and
move on.

## 1. Find (or create) the compose project

- Locate the compose file that defines the development environment. It may be at
  the repo root (`compose.yaml`, `compose.yml`, `docker-compose.yml`,
  `docker-compose.yaml`) or, in devcontainer-style projects, under
  `.devcontainer/`.
- Record its path from the repository root; you will need it in step 2.
- If the project has **no** compose file, create one at `.devcontainer/compose.yml`
  that builds the app image, bind-mounts the source tree, and declares the services
  the app needs (database, cache, etc.). Use the pattern in step 3.
- Note whether the compose file has a top-level `name:` or its containers already
  run under a project name — you will reuse that name in step 2.

## 2. Pin the project name and the compose file

Create or append `.env` at the repository root:

```dotenv
COMPOSE_PROJECT_NAME=<project-name>
COMPOSE_FILE=<path-to-compose-file-relative-to-repo-root>
```

Rules:

- `<project-name>` must contain only lowercase letters, digits, `-` and `_`, and
  must start with a letter or digit. Reuse the name the project already runs
  under if there is one.
- `COMPOSE_FILE` is only strictly required when the compose file is not
  default-discovered (for example `.devcontainer/compose.yml`). Setting it
  explicitly is harmless and recommended.
- `.env` must be either committed or gitignored. A committed `.env` should not
  contain secrets; if it would, keep it gitignored and list it in
  `.worktreeinclude` (step 5).

Why this matters: `composefork` names each fork `<project-name>-<worktree-dir>`
and uses the pinned parent name to find cached volumes. Without `.env`, every
worktree infers its own project name and the volume cache silently stops working.

## 3. Make dependencies cacheable

The point of the tool is that a fork starts from a snapshot of installed
dependencies instead of reinstalling them. For every service that installs
dependencies at container start (an entrypoint running `bundle install`,
`npm ci`, `pip install`, etc.):

- Move each dependency directory into its own **top-level named volume** and mount
  it at the right path. Named volumes are the only thing `composefork cache` can
  snapshot.
- **Bind-mount the source tree** so each worktree builds and runs its own code.
- Add a **healthcheck** that only passes once dependency installation has
  finished and the service is actually ready. `composefork up` waits for health,
  and `composefork cache` snapshots volumes after that point; without a
  healthcheck the snapshot can be taken mid-install.
- Do **not** give a volume an explicit `name:` or `external: true`. composefork
  matches cache tarballs by the compose-prefixed `<project>_<key>` name, and an
  explicit name both breaks the cache and risks a fork mounting the main
  checkout's live volume.

A minimal pattern, for a compose file at the repo root:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: .devcontainer/Dockerfile
    working_dir: /app
    volumes:
      - .:/app                  # source bind mount: one per worktree
      - deps:/app/node_modules  # installed dependencies: snapshotted and reused
    ports:
      - "127.0.0.1:3000:3000"   # only used by the main checkout; forks get a dynamic loopback port
    depends_on:
      - db
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:3000/up"]
      interval: 5s
      timeout: 5s
      retries: 3
      start_period: 120s
  db:
    image: postgres:17
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  deps:
  db_data:
```

If the compose file lives in `.devcontainer/`, the build context and bind mount
are usually `..` rather than `.`.

## 4. Ports and networking

- `composefork` discards authored host ports in forks and rebinds every published
  port to `127.0.0.1` on an ephemeral port, so forks cannot collide. Do not make
  the project depend on a fixed host port (OAuth callbacks, hardcoded URLs);
  `composefork ps` is how a port is discovered.
- Publish only what the host needs; services on the same compose network reach
  each other without publishing anything. Do not republish a port the developer
  already runs locally (for example a rbenv Postgres on 5432) — it collides with
  their service. When a port must be published for the main checkout, let the
  host port be assigned rather than pinned, so it cannot clash.
- Keep services on normal bridge networks. Do **not** use `network_mode: host` or
  `network_mode: service:...`, which bypass the fork's isolation.
- If the compose file references `external: true` networks, make sure those
  networks actually exist before `up`, and document how to create them. A fork
  will not create them.

## 5. Seed local files into worktrees

A linked worktree only checks out tracked files, so gitignored files the project
needs at runtime (`.env`, a local database, credentials, uploaded files) are
missing. List them in `.worktreeinclude` at the repository root:

```
.env
storage/
*.local
config/master.key
```

Two things populate them, which is the part agents find confusing:

- **Claude Code copies every `.worktreeinclude` file automatically** when it
  creates a worktree with its worktree option, so those checkouts already have
  them.
- **`composefork up` covers the rest.** A plain `git worktree add` does not copy
  them, so on every start `up` resolves the list and copies any missing file from
  the main checkout into the fork. It fills in only what is absent and never
  overwrites edits the fork has made.

Rules:

- Entries are gitignore-style patterns.
- Every listed path must be **gitignored** (untracked), or already tracked. An
  untracked file that is not gitignored cannot be copied. In particular, if
  `.env` is gitignored it must be listed here, or forks will start without it.

## 6. Tell agents how to use composefork

Add a section to `AGENTS.md`, `CLAUDE.md`, or the equivalent file your agent
reads. Adapt the wording, but keep the substance:

```markdown
## Container environment

Use `composefork`, not `docker compose`, to run this project in a worktree.

- `composefork up` — start this worktree's isolated environment
- `composefork ps` — services, health, and their dynamically assigned ports
- `composefork exec <service> <command...>` — run a command in a service
- `composefork down` — tear the environment down when you are done

Always run these from the root of the worktree. Read `composefork help` for the
full reference, or `composefork help <command>` for one command. The main
checkout is the human's own environment, managed with plain `docker compose`;
composefork refuses to run there.
```

If your agent supports lifecycle hooks, add one that runs `composefork down`
when a session ends, so finished worktrees do not leave containers behind. In
Claude Code, merge this into `.claude/settings.json`; adapt the equivalent for
other agents:

```json
{
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "command -v composefork >/dev/null 2>&1 && composefork down || true"
          }
        ]
      }
    ]
  }
}
```

## 7. Build the volume cache

Run `composefork cache` from the main checkout **before** agents start working.
It builds and starts a throwaway copy of the project under a random name, waits
for the healthchecks added in step 3 to go green, snapshots the named volumes
(now holding the installed dependencies), and tears the copy down. It leaves the
developer's running containers and data untouched.

This is what makes the first `composefork up` in a worktree fast. Without a
cache, every fork installs its dependencies from scratch the first time, which is
slow and is the usual reason a fork appears to hang. The cache exists only if
step 3 put the dependencies in named volumes and gave the services a healthcheck
that turns green once the install finishes; `composefork up` still works without
it, just slower.

Re-run `composefork cache` whenever dependencies change (a lockfile or manifest
edit). The cache belongs to the project, not to any one worktree.

## 8. Verify

Commit the new files before this step. A linked worktree only checks out
committed files, so an uncommitted `.devcontainer/`, `.worktreeinclude` or
`.env` never reaches it; if you cannot commit yet, copy those files into the
worktree by hand first.

1. Create a throwaway linked worktree (`git worktree add ../cf-smoke-test`) and,
   from inside it, run `composefork up`, then `composefork ps`,
   `composefork exec <service> true`, then `composefork down`. `up` should report
   restoring cached volumes and reach health quickly.
2. Confirm the main checkout still comes up under plain `docker compose up`.
3. Report anything you could not satisfy, especially the items in the
   assumptions list below.

---

## Assumptions composefork makes about your project

Keep this as a checklist; anything unmet should be surfaced to the developer
rather than silently accepted.

### Git and worktrees

- The project is a **git repository** and `git` is on `PATH`.
- The main checkout is a normal (non-bare) worktree. Bare repositories are not
  supported.
- Forks must be **linked worktrees of the same repository** created with
  `git worktree add`. A separate clone is indistinguishable from the main
  checkout and is refused.
- Commands must run from the **worktree root**. `.env` is loaded from the current
  directory, and `COMPOSE_FILE` is resolved relative to it.
- The **worktree directory name becomes part of the compose project name**
  (`<parent>-<dirname>`).

### Compose project

- The compose project is discoverable from the working directory, via
  `COMPOSE_FILE` or a default-named file at the root or an ancestor.
  `.devcontainer/compose.yml` is **not** default-discovered and needs
  `COMPOSE_FILE`.
- `COMPOSE_PROJECT_NAME` is pinned and already valid (lowercase, `[a-z0-9_-]`,
  not starting with `-` or `_`).
- Compose files are local regular files; there is no support for remote or stdin
  configs.
- Interpolation uses only the OS environment and `.env`; composefork adds no
  defaults of its own.

### Volumes and the cache

- Only **top-level named volumes** are cached. Bind mounts and anonymous volumes
  are not.
- Volumes use the default `<project>_<key>` naming. Explicit `name:` and
  `external: true` volumes are cached but never restored, and a fork may mount the
  main checkout's live volume.
- Volumes are plain file trees that can be tarred and restored; exotic volume
  drivers are untested.
- `cache` snapshots **all** named volumes of a throwaway copy of the project
  under a random name, including data volumes such as the database, in their
  freshly initialised state — not the developer's current data. Forks import
  that snapshot.
- Services that install dependencies on start expose a **healthcheck** that only
  passes once they are ready. Without one, `up` still works but the snapshot may
  be incomplete.
- `cache` runs a throwaway project under a random name with loopback, dynamic
  ports, so it can run without stopping the developer's stack.

### Services, ports and networks

- Services use bridge networking; `network_mode: host` or `service:` defeats
  isolation.
- Host port publishing is rewritten in forks: authored host ports are dropped and
  republished on loopback with ephemeral ports. Stable host ports are not
  available to forks.
- `up` always builds the project; build and startup must succeed under the fork's
  project name.
- Services that pin an explicit `image:` tag share that image across forks; only
  `build:` services get per-fork images.
- External networks must already exist; composefork never creates them.

### Runtime environment

- A reachable Docker daemon and Docker Compose v5 semantics.
- A Linux-like host: the tool uses `flock`, `os.UserCacheDir()` for its snapshot
  cache, and a scratch container that runs as host UID. Rootless podman is
  untested.
