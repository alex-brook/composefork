- It wasn't obvious that `composefork worktree up` waits for the project to be healthy
    - [ ] Show health as a column in project info

- [] When a project container died, the agent got stuck
    - [] Add `composefork worktree restart` command

- [x] The agent confused ls and ps, is there misleading docs?

- [x] There is no exec command, the agent has to construct a vanilla compose command with the project name

- [] Volumes are not copied on fork, which means they take a long time to start
    - [] Dump to tmp file
    - [] Restore to new project

- Investigate checkpoints to avoid using too much RAM
    - [] Guess memory consumption based on avg. of existing projects
    - [] Decide if we will go over the allocated docker RAM with a new fork
    - [] Checkpoint/pause the oldest
    - [] When you interact with a paused project, prints a note for agent to ask permission before unpausing
