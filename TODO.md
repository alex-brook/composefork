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
    - [x] System container for these kinds of operations
    - [x] Snapshot volumes
    - [x] Use these cached volumes when creating forks

- [] Running `composefork up` in the main worktree should be equivalent to docker compose up

- [] Each fork should build its own image, not just use the parent image
    - [] Fork images should be removed when the project is torn down

- [] Setup prompt that covers:
    - [] That a compose project exists
    - [] That it can be recognised from the root directory (.env)
    - [] That healthchecks are defined for every service that installs deps on start
    - [] Adding a claude hook that runs `composefork worktree down` on SessionEnd if composefork is installed
    - [] Adding context to AGENTS.md or CLAUDE.md about composefork, and how to use `composefork skill` to get further context

- [] Caching should be invisible to the user
    - [] First time agent setup should take an initial cache
    - [] Some kind of recurring cache behaviour that is invisible

- Investigate checkpoints to avoid using too much RAM
    - [] Guess memory consumption based on avg. of existing projects
    - [] Decide if we will go over the allocated docker RAM with a new fork
    - [] Checkpoint/pause the oldest
    - [] When you interact with a paused project, prints a note for agent to ask permission before unpausing
