// Package git provides git subprocess wrappers and file copy helpers.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoRoot returns the absolute path to the git repository root.
func RepoRoot() (string, error) {
	out, err := run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(out), nil
}

// WorktreeAdd creates a new worktree at path. When createBranch is true a new
// branch named branch is created from HEAD; otherwise an existing branch is
// checked out.
func WorktreeAdd(path, branch string, createBranch bool) error {
	var args []string
	if createBranch {
		// git worktree add -b <branch> <path>  — branch from HEAD
		args = []string{"worktree", "add", "-b", branch, path}
	} else {
		// git worktree add <path> <branch>  — check out existing branch
		args = []string{"worktree", "add", path, branch}
	}

	//nolint:gosec // args are constructed from validated git subcommands and user-supplied branch/path
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WorktreeRemove removes the worktree at path.
func WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeleteBranch deletes a local branch. Use force=true to delete even if
// unmerged (equivalent to git branch -D).
func DeleteBranch(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := run("git", "branch", flag, branch)
	return err
}

// Worktree holds info about a single git worktree.
type Worktree struct {
	Path   string
	Branch string
	HEAD   string
	IsMain bool
}

// ListWorktrees returns all worktrees for the current repo.
func ListWorktrees() ([]Worktree, error) {
	out, err := run("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out), nil
}

func parseWorktreeList(raw string) []Worktree {
	var trees []Worktree
	var current Worktree
	first := true

	for line := range strings.SplitSeq(raw, "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			if current.Path != "" {
				trees = append(trees, current)
				current = Worktree{}
			}
			continue
		}

		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			current.Path = path
			current.IsMain = first
			first = false
		} else if head, ok := strings.CutPrefix(line, "HEAD "); ok {
			current.HEAD = head[:min(8, len(head))]
		} else if branch, ok := strings.CutPrefix(line, "branch "); ok {
			current.Branch, _ = strings.CutPrefix(branch, "refs/heads/")
		} else if line == "detached" {
			current.Branch = "(detached)"
		}
	}

	if current.Path != "" {
		trees = append(trees, current)
	}

	return trees
}

// CopyPath copies src (file or directory) into dst directory.
// If src does not exist it is silently skipped.
func CopyPath(src, dstDir string) error {
	info, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	dst := filepath.Join(dstDir, filepath.Base(src))
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

// CopyRelativePath copies repoRoot/relPath into dstWorktree, preserving the
// sub-path structure. Missing source paths are silently skipped.
//
// Example: repoRoot=~/proj, relPath=".serena/project.yml", dstWorktree=~/proj-feat
// copies to ~/proj-feat/.serena/project.yml.
func CopyRelativePath(repoRoot, relPath, dstWorktree string) error {
	src := filepath.Join(repoRoot, filepath.Clean(relPath))
	info, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	dst := filepath.Join(dstWorktree, filepath.Clean(relPath))
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	//nolint:gosec // src is a trusted path from within the repository
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, _ := os.Stat(src)
	return os.WriteFile(dst, data, info.Mode())
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func run(name string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	//nolint:gosec // name is always "git"; args are constructed from safe subcommands
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// MinGitMajor and MinGitMinor define the minimum supported git version.
const (
	MinGitMajor = 2
	MinGitMinor = 28
)

// CheckVersion returns an error if the installed git version is older than
// 2.28, or if git is not found in PATH.
func CheckVersion() error {
	out, err := run("git", "--version")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	major, minor, err := parseGitVersion(strings.TrimSpace(out))
	if err != nil {
		return fmt.Errorf("could not parse git version %q: %w", out, err)
	}

	if major < MinGitMajor || (major == MinGitMajor && minor < MinGitMinor) {
		return fmt.Errorf(
			"wt requires git >= %d.%d, found %d.%d — please upgrade git",
			MinGitMajor, MinGitMinor, major, minor,
		)
	}

	return nil
}

// parseGitVersion extracts major and minor from "git version X.Y.Z[...]".
func parseGitVersion(s string) (major, minor int, err error) {
	// Expected format: "git version 2.39.1" or "git version 2.39.1.windows.1"
	s = strings.TrimPrefix(s, "git version ")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected format: %q", s)
	}

	if _, err := fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return 0, 0, fmt.Errorf("parse major %q: %w", parts[0], err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &minor); err != nil {
		return 0, 0, fmt.Errorf("parse minor %q: %w", parts[1], err)
	}

	return major, minor, nil
}
