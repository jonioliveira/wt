package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jonioliveira/wt/internal/config"
	"github.com/jonioliveira/wt/internal/git"
	"github.com/jonioliveira/wt/internal/ui"
)

func newNewCmd() *cobra.Command {
	var (
		path     string
		noBranch bool
	)

	cmd := &cobra.Command{
		Use:   "new <branch>",
		Short: "Create a new worktree and sync AI context files",
		Long: `Creates a git worktree for <branch> and copies configured context files
(e.g. .claude/, .serena/project.yml, CLAUDE.md) from the repo root into it.

By default a new branch is created. Use --no-branch to check out an existing one.

Examples:
  wt new feature/my-feature
  wt new feature/my-feature --path .worktrees/my-feature
  wt new main --no-branch`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(cmd, args[0], path, noBranch)
		},
	}

	cmd.Flags().StringVar(&path, "path", "",
		"custom path for the worktree (default: .worktrees/<branch>)")
	cmd.Flags().BoolVar(&noBranch, "no-branch", false,
		"check out an existing branch instead of creating a new one")

	return cmd
}

func runNew(cmd *cobra.Command, branch, path string, noBranch bool) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return err
	}

	worktreePath := path
	if worktreePath == "" {
		safeBranch := strings.ReplaceAll(branch, "/", "-")
		worktreePath = filepath.Join(repoRoot, ".worktrees", safeBranch)
	}

	absPath, err := filepath.Abs(worktreePath)
	if err != nil {
		return err
	}

	cfg, err := config.Load(repoRoot)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	_, _ = ui.Bold.Printf("Creating worktree: %s\n", absPath)
	fmt.Printf("  Branch: %s\n\n", branch)

	if err := git.WorktreeAdd(absPath, branch, !noBranch); err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}

	fmt.Println()
	_, _ = ui.Bold.Println("Syncing context files...")

	result := syncContext(repoRoot, absPath, cfg.Copy)
	result.PrintSummary(cmd.OutOrStdout())
	fmt.Printf("\n  cd %s\n\n", absPath)

	return nil
}
