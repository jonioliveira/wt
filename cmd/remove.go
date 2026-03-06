package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jonioliveira/wt/internal/git"
	"github.com/jonioliveira/wt/internal/ui"
)

func newRemoveCmd() *cobra.Command {
	var force, keepBranch bool

	cmd := &cobra.Command{
		Use:     "rm <path-or-branch>",
		Aliases: []string{"remove"},
		Short:   "Remove a worktree",
		Long: `Removes a git worktree by path or branch name.

Examples:
  wt rm feature/my-feature
  wt rm feature/my-feature --force
  wt rm feature/my-feature --keep-branch`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			target := args[0]
			path, branch, err := resolveWorktree(target)
			if err != nil {
				return err
			}

			fmt.Printf("Removing worktree: %s\n", path)

			if err := git.WorktreeRemove(path, force); err != nil {
				return fmt.Errorf("git worktree remove: %w", err)
			}

			if branch != "" && !keepBranch {
				if err := git.DeleteBranch(branch, force); err != nil {
					_, _ = ui.Yellow.Printf(
						"⚠  worktree removed but could not delete branch %q: %v\n",
						branch,
						err,
					)
				}
			}

			_, _ = ui.Green.Println("✓ Worktree removed")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false,
		"force removal even if the worktree has uncommitted changes")
	cmd.Flags().BoolVar(&keepBranch, "keep-branch", false,
		"keep the branch after removing the worktree")

	return cmd
}
