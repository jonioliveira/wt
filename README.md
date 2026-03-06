# wt

> Git worktrees with AI context — create, sync, and remove worktrees without losing your assistant's brain.

Git worktrees are great for working on multiple branches in parallel. But your AI context doesn't travel with them. `wt` fixes that.

When you create a worktree with `wt new`, it automatically copies your configured context files (`.claude/`, `.serena/project.yml`, `CLAUDE.md`, or anything else you configure) into the new worktree so your AI assistant is immediately ready to work.

---

## Requirements

- **git >= 2.28** (released July 2020)
- **Go 1.22+** — for building from source

## Install

**Homebrew** (recommended):
```bash
brew tap jonioliveira/tap
brew install wt
```

**Go**:
```bash
go install github.com/jonioliveira/wt@latest
```

Or download a binary from [Releases](https://github.com/jonioliveira/wt/releases).

---

## Usage

```bash
# Create a worktree on a new branch
wt new feature/my-feature

# Create at a custom path
wt new feature/my-feature --path .worktrees/my-feature

# Check out an existing branch instead of creating a new one
wt new main --no-branch

# List all worktrees with context file status
wt ls

# Re-sync context files into an existing worktree
wt sync feature/my-feature
wt sync                        # syncs into current worktree directory

# Remove a worktree and its branch
wt rm feature/my-feature
wt rm feature/my-feature --force        # force if uncommitted changes
wt rm feature/my-feature --keep-branch  # remove worktree but keep the branch

# Scaffold a .wtconfig.yml in the current repo
wt init
```

---

## Configuration

By default, `wt` copies these paths:

```
.claude/
.serena/project.yml
CLAUDE.md
```

Run `wt init` to generate a `.wtconfig.yml` you can customise:

```yaml
# .wtconfig.yml
copy:
  - .claude/
  - .serena/project.yml
  - CLAUDE.md
  # - .cursor/rules     # Cursor
  # - .copilot/         # GitHub Copilot
  # - .env.example
```

Paths that don't exist in your repo are silently skipped.

---

## How it works

`wt new <branch>` is equivalent to:

```bash
git worktree add -b <branch> .worktrees/<branch>
cp -r .claude/            .worktrees/<branch>/
cp -r .serena/project.yml .worktrees/<branch>/.serena/
cp    CLAUDE.md           .worktrees/<branch>/
```

Worktrees are created at `.worktrees/<branch>` inside the repo root (slashes in branch names become dashes). Override with `--path`.

---

## Why not just commit these files?

You might — and if `.claude/` and `CLAUDE.md` are committed, git already puts them in every worktree. `wt` is for the case where these files are gitignored: team-specific agent configs, `.serena/memories/`, or personal `CLAUDE.md` preferences you don't want in the repo.

---

## Claude Code skills

The `skills/` directory contains ready-to-use Claude Code agents. Copy them to your `.claude/agents/` to use them:

- **`wt-onboard`** — orients a new Claude session to the `wt` codebase (reads structure, conventions, git status)
- **`wt-feature`** — guides the full feature lifecycle: `wt new` → work → `wt sync` → `wt rm`

---

## Contributing

PRs welcome. The project is intentionally small — the goal is to stay a thin, focused tool. See [CONTRIBUTING](.github/pull_request_template.md) for the PR checklist.

```
wt/
├── cmd/               # cobra commands (new, rm, ls, sync, init)
├── internal/
│   ├── config/        # .wtconfig.yml loading
│   ├── git/           # git worktree wrappers + file copy helpers
│   └── ui/            # terminal output helpers
├── skills/            # Claude Code agent skills
└── main.go
```

---

## License

MIT
