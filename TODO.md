- It wasn't obvious that `composefork worktree up` waits for the project to be healthy
    - [x] Show health as a column in project info

- [x] When a project container died, the agent got stuck
    - [x] Add `composefork worktree restart` command

- [x] The agent confused ls and ps, is there misleading docs?

- [x] There is no exec command, the agent has to construct a vanilla compose command with the project name

- [] Add version command

- [] Volumes are not copied on fork, which means they take a long time to start
    - [x] Add a new command `composefork cache`
    - [x] Stop the parent project
    - [x] System container for these kinds of operations
    - [x] Snapshot volumes
    - [] Use these cached volumes when creating forks

- Investigate checkpoints to avoid using too much RAM
    - [] Guess memory consumption based on avg. of existing projects
    - [] Decide if we will go over the allocated docker RAM with a new fork
    - [] Checkpoint/pause the oldest
    - [] When you interact with a paused project, prints a note for agent to ask permission before unpausing
