## What

<!-- What does this PR do? One or two sentences. -->

## Why

<!-- Why is this change needed? Link to an issue if applicable. -->

## How

<!-- Brief description of the approach, especially if non-obvious. -->

## Checklist

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] `golangci-lint run` passes
- [ ] New subcommands follow the constructor pattern in `CLAUDE.md`
- [ ] Error messages use `fmt.Errorf("verb noun: %w", err)` format
- [ ] No interactive prompts, network calls, or state outside git/`.wtconfig.yml`
