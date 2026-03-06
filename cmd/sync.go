package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonioliveira/wt/internal/config"
	"github.com/jonioliveira/wt/internal/git"
	"github.com/jonioliveira/wt/internal/ui"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [path-or-branch]",
		Short: "Re-sync context files into a worktree",
		Long: `Copies configured context files from the repo root into an existing worktree.
Useful when you've updated your .claude/ agents or CLAUDE.md and want to propagate changes.

If no target is given, syncs into the current directory (must be inside a worktree).

Examples:
  wt sync                       # sync into current worktree
  wt sync ../myrepo-feature
  wt sync feature/my-feature`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				return err
			}

			target, err := resolveTarget(args, repoRoot)
			if err != nil {
				return err
			}

			if _, err := os.Stat(target); os.IsNotExist(err) {
				return fmt.Errorf("worktree path does not exist: %s", target)
			}

			cfg, err := config.Load(repoRoot)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			_, _ = ui.Bold.Printf("Syncing context → %s\n\n", target)

			result := syncContext(repoRoot, target, cfg.Copy)
			result.PrintSummary(cmd.OutOrStdout())

			return nil
		},
	}
}

// resolveTarget returns the absolute worktree path to sync into.
// With no args it uses the current directory; with one arg it resolves
// branch names to paths via the live worktree list.
func resolveTarget(args []string, repoRoot string) (string, error) {
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		if cwd == repoRoot {
			return "", fmt.Errorf(
				"already in the main worktree root; specify a target branch or path")
		}
		return cwd, nil
	}

	return resolveWorktreePath(args[0])
}
