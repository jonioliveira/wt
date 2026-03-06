# wt — Claude Context

## What this project is

`wt` is a small Go CLI that wraps `git worktree` to automatically sync AI context files (`.claude/`, `.serena/project.yml`, `CLAUDE.md`, etc.) into new worktrees. The goal is to stay small and focused — a thin, composable tool that does one thing well.

## Project structure

```
wt/
├── main.go                   # calls cmd.NewRootCmd().Execute(), owns os.Exit
├── cmd/
│   ├── root.go               # NewRootCmd() — wires all subcommands, no package-level state
│   ├── context.go            # shared helpers: syncContext(), resolveWorktreePath(), resolveTarget()
│   ├── new.go                # newNewCmd() *cobra.Command
│   ├── remove.go             # newRemoveCmd() *cobra.Command
│   ├── list.go               # newListCmd() *cobra.Command
│   ├── sync.go               # newSyncCmd() *cobra.Command
│   └── init.go               # newInitCmd() *cobra.Command
├── internal/
│   ├── config/
│   │   └── config.go         # .wtconfig.yml loading via koanf; always returns a valid *Config
│   ├── git/
│   │   └── git.go            # git subprocess wrappers + file copy helpers
│   └── ui/
│       └── printer.go        # shared color vars (Bold, Green, Yellow, Cyan, Dim) + SyncResult
├── .wtconfig.yml.example
├── .golangci.yml
├── .goreleaser.yml
└── .github/
    └── workflows/
        ├── ci.yml             # lint + test + build on push/PR
        └── release.yml        # GoReleaser on git tag
```

## Requirements

- **git >= 2.28** (July 2020) — required for `git init -b` and reliable worktree support
- **Go 1.22+**

## Stack

- **Go 1.22**
- **cobra** — CLI framework
- **koanf** — config file loading (explicit, no globals, composable providers)
- **fatih/color** — terminal output coloring
- **GoReleaser** — cross-platform binary releases via GitHub Actions

## Key conventions

### Command construction
Every subcommand is a constructor function (`newNewCmd() *cobra.Command`), not a package-level var. Flags are declared inside the constructor as local vars — no shared mutable state between commands. All wiring happens in `NewRootCmd()` in `root.go`.

```go
// correct
func newNewCmd() *cobra.Command {
    var path string
    cmd := &cobra.Command{...}
    cmd.Flags().StringVar(&path, "path", "", "...")
    return cmd
}

// never do this
var newCmd = &cobra.Command{...}
func init() { rootCmd.AddCommand(newCmd) }
```

### Shared logic lives in cmd/context.go
`syncContext()` is the single implementation of the copy loop, used by both `new` and `sync`. `resolveWorktreePath()` handles branch-name → path resolution. Never duplicate this logic in individual commands.

### internal/ui for all terminal output
Import `ui.Bold`, `ui.Green`, etc. from `internal/ui`. Never construct `color.New(...)` inline in command files. `ui.SyncResult` carries copy/skip/fail counts and knows how to print its own summary.

### internal/config contract
`config.Load()` always returns a valid `*Config` — never nil, never an error for a missing file. Missing config → defaults. Empty `Copy` slice → defaults. Callers never need to guard against nil.

### internal/git contract
`git.CopyRelativePath()` silently skips missing source paths. Callers in `syncContext()` do their own `os.Stat` check first to give user-visible feedback before attempting the copy.

### Error wrapping
Use `fmt.Errorf("verb noun: %w", err)` — lowercase, no trailing period, verb-noun ordering:
```go
return fmt.Errorf("load config: %w", err)      // correct
return fmt.Errorf("Failed to load config: %w", err)  // wrong
```

### cobra settings
`SilenceUsage: true` and `SilenceErrors: true` are set on the root command. Usage is only shown when it's actually helpful (e.g. wrong number of args), not on every runtime error. `main.go` owns error printing and `os.Exit`.


## Running locally

```bash
go mod tidy
go build -o wt .
./wt --help
./wt new feature/test-branch
```

## Linting

Uses **golangci-lint v2**. Config in `.golangci.yml`.

```bash
golangci-lint run
```

CI runs lint + test + build on every push and PR via `.github/workflows/ci.yml`.

## Release process

```bash
git tag v0.x.0
git push --tags
```

GoReleaser builds linux/darwin/windows × amd64/arm64. Config in `.goreleaser.yml`.

## Adding a new subcommand

1. Create `cmd/<name>.go` with `func new<Name>Cmd() *cobra.Command`
2. Declare flags as local vars inside the constructor
3. Register in `NewRootCmd()` in `root.go`
4. Delegate any shared file/git logic to `internal/`

## What to keep out of scope

- No interactive prompts — all input via args and flags
- No network calls — purely a local git tool
- No shell execution other than `git` subprocesses
- No state storage outside `.wtconfig.yml` and git itself
