package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jonioliveira/wt/internal/config"
	"github.com/jonioliveira/wt/internal/git"
	"github.com/jonioliveira/wt/internal/ui"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all worktrees with context file status",
		Long:    `Lists all worktrees for the current repository and shows which AI context files are present in each.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				return err
			}

			trees, err := git.ListWorktrees()
			if err != nil {
				return fmt.Errorf("list worktrees: %w", err)
			}

			cfg, err := config.Load(repoRoot)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			for i, t := range trees {
				if i > 0 {
					fmt.Println()
				}
				printWorktree(t, cfg.Copy)
			}

			return nil
		},
	}
}

func printWorktree(t git.Worktree, paths []string) {
	label := ""
	if t.IsMain {
		label = " (main)"
	}

	_, _ = ui.Bold.Printf("%s", t.Path)
	_, _ = ui.Cyan.Printf("  %s", t.Branch)
	_, _ = ui.Dim.Printf("  %s%s\n", t.HEAD, label)

	for _, relPath := range paths {
		src := filepath.Join(t.Path, relPath)
		if _, err := os.Stat(src); err == nil {
			_, _ = ui.Green.Printf("  ✓  %s\n", relPath)
		} else {
			_, _ = ui.Yellow.Printf("  ✗  %s\n", relPath)
		}
	}
}
