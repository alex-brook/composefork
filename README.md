# Composefork

Let your agents reuse your docker compose devcontainer

`composefork` is a small utility that lets you use your existing docker compose
based devcontainer configuration with parallel agents. It works by cloning the
original compose project and namespacing it for a single worktree

`composefork` complements a traditional development workflow. The original
worktree and compose project is reserved for you to work on manually.

## Prerequisites
- Docker
- Git
- A Harness

## Installation
### Mise
```
mise use -g github:alex-brook/composefork@latest
```
### Manual
Download a binary matching your system from the releases page and add it to your PATH

## Getting started
- Run `composefork setup | claude -p` or similar in your project root
