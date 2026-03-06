# Changelog

## v1.0.0 — Initial Release

### Commands

- `wt new <branch>` — create a git worktree on a new branch and automatically sync AI context files into it. Defaults to `.worktrees/<branch>` inside the repo root. Supports `--path` for a custom location and `--no-branch` to check out an existing branch.
- `wt rm <branch>` — remove a worktree and delete its local branch. Supports `--force` for uncommitted changes and `--keep-branch` to preserve the branch.
- `wt sync [branch]` — re-sync context files into an existing worktree. Runs from inside a worktree with no arguments, or targets a branch/path explicitly.
- `wt ls` — list all worktrees with branch, HEAD, and context file status.
- `wt init` — scaffold a `.wtconfig.yml` with sensible defaults in the current repo.

### Configuration

- `.wtconfig.yml` — declare which files and directories to copy into each new worktree. Missing paths are silently skipped. Defaults to `.claude/`, `.serena/project.yml`, and `CLAUDE.md` when no config file is present.

### Claude Code skills

- `skills/wt-onboard.md` — agent that orients a new Claude session to the `wt` codebase.
- `skills/wt-feature.md` — agent that guides the full feature worktree lifecycle.
