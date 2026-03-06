package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonioliveira/wt/internal/git"
	"github.com/jonioliveira/wt/internal/ui"
)

// syncContext copies each path in paths from repoRoot into destWorktree,
// printing status for each entry. Missing source paths are silently skipped.
func syncContext(repoRoot, destWorktree string, paths []string) ui.SyncResult {
	var result ui.SyncResult

	for _, relPath := range paths {
		relPath = strings.TrimSuffix(relPath, "/")
		src := filepath.Join(repoRoot, relPath)

		if _, err := os.Stat(src); os.IsNotExist(err) {
			_, _ = ui.Dim.Printf("  ⚠  skipped  %-38s (not found)\n", relPath)
			result.Skipped++
			continue
		}

		if err := git.CopyRelativePath(repoRoot, relPath, destWorktree); err != nil {
			_, _ = ui.Yellow.Printf("  ✗  failed   %-38s %v\n", relPath, err)
			result.Failed++
			continue
		}

		_, _ = ui.Green.Printf("  ✓  copied   %s\n", relPath)
		result.Copied++
	}

	fmt.Println()
	return result
}

// resolveWorktreePath returns the absolute filesystem path for target, which
// may be either a branch name or a filesystem path. It resolves branch names
// by scanning the live worktree list.
func resolveWorktreePath(target string) (string, error) {
	path, _, err := resolveWorktree(target)
	return path, err
}

// resolveWorktree returns the absolute path and branch name for target.
// target may be a branch name or a filesystem path.
func resolveWorktree(target string) (path, branch string, err error) {
	trees, listErr := git.ListWorktrees()
	if listErr == nil {
		for _, t := range trees {
			if t.Branch == target {
				return t.Path, t.Branch, nil
			}
		}
		// target may be a path — find matching branch
		abs, absErr := filepath.Abs(target)
		if absErr != nil {
			return "", "", absErr
		}
		for _, t := range trees {
			if t.Path == abs {
				return t.Path, t.Branch, nil
			}
		}
		return abs, "", nil
	}

	abs, absErr := filepath.Abs(target)
	return abs, "", absErr
}
