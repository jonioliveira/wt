package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jonioliveira/wt/internal/git"
)

// ── CopyPath ─────────────────────────────────────────────────────────────────

func TestCopyPath_File(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "CLAUDE.md"), "# context")

	if err := git.CopyPath(filepath.Join(src, "CLAUDE.md"), dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(dst, "CLAUDE.md"), "# context")
}

func TestCopyPath_Directory(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, ".claude", "settings.json"), `{"key":"val"}`)
	writeFile(t, filepath.Join(src, ".claude", "agents", "coder.md"), "# coder")

	if err := git.CopyPath(filepath.Join(src, ".claude"), dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(dst, ".claude", "settings.json"), `{"key":"val"}`)
	assertFileContent(t, filepath.Join(dst, ".claude", "agents", "coder.md"), "# coder")
}

func TestCopyPath_MissingSource_IsNoOp(t *testing.T) {
	dst := t.TempDir()

	if err := git.CopyPath("/nonexistent/path/file.txt", dst); err != nil {
		t.Errorf("expected nil for missing source, got: %v", err)
	}
}

func TestCopyPath_PreservesFileMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFileMode(t, filepath.Join(src, "script.sh"), "#!/bin/bash", 0o755)

	if err := git.CopyPath(filepath.Join(src, "script.sh"), dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dst, "script.sh"))
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode() != 0o755 {
		t.Errorf("want mode 0755, got %v", info.Mode())
	}
}

// ── CopyRelativePath ─────────────────────────────────────────────────────────

func TestCopyRelativePath_File(t *testing.T) {
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, "CLAUDE.md"), "# hello")

	if err := git.CopyRelativePath(repo, "CLAUDE.md", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, "CLAUDE.md"), "# hello")
}

func TestCopyRelativePath_NestedFile(t *testing.T) {
	// .serena/project.yml must land at worktree/.serena/project.yml with the
	// intermediate directory created automatically.
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, ".serena", "project.yml"), "language: ruby")

	if err := git.CopyRelativePath(repo, ".serena/project.yml", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, ".serena", "project.yml"), "language: ruby")
}

func TestCopyRelativePath_Directory(t *testing.T) {
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, ".claude", "settings.json"), `{}`)
	writeFile(t, filepath.Join(repo, ".claude", "agents", "reviewer.md"), "# reviewer")

	if err := git.CopyRelativePath(repo, ".claude", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, ".claude", "settings.json"), `{}`)
	assertFileContent(t, filepath.Join(worktree, ".claude", "agents", "reviewer.md"), "# reviewer")
}

func TestCopyRelativePath_MissingSource_IsNoOp(t *testing.T) {
	repo, worktree := t.TempDir(), t.TempDir()

	if err := git.CopyRelativePath(repo, ".claude", worktree); err != nil {
		t.Errorf("expected nil for missing source, got: %v", err)
	}

	entries, _ := os.ReadDir(worktree)
	if len(entries) != 0 {
		t.Errorf("worktree should be empty, found %d entries", len(entries))
	}
}

func TestCopyRelativePath_TrailingSlash(t *testing.T) {
	// filepath.Clean inside the impl strips trailing slashes; the destination
	// must not gain a double-nested directory.
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, ".claude", "settings.json"), `{}`)

	if err := git.CopyRelativePath(repo, ".claude/", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, ".claude", "settings.json"), `{}`)
}

func TestCopyRelativePath_OverwritesExistingFile(t *testing.T) {
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, "CLAUDE.md"), "new content")
	writeFile(t, filepath.Join(worktree, "CLAUDE.md"), "old content")

	if err := git.CopyRelativePath(repo, "CLAUDE.md", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, "CLAUDE.md"), "new content")
}

func TestCopyRelativePath_DeepDirectoryTree(t *testing.T) {
	repo, worktree := t.TempDir(), t.TempDir()

	writeFile(t, filepath.Join(repo, ".claude", "agents", "deep", "nested.md"), "deep")

	if err := git.CopyRelativePath(repo, ".claude", worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFileContent(t, filepath.Join(worktree, ".claude", "agents", "deep", "nested.md"), "deep")
}

// ── RepoRoot ──────────────────────────────────────────────────────────────────

func TestRepoRoot_InsideRepo(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	root, err := git.RepoRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resolve symlinks on both sides: macOS TempDir resolves
	// /var/folders/... → /private/var/folders/...
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(dir=%q): %v", dir, err)
	}
	got, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks(root=%q): %v", root, err)
	}
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestRepoRoot_InsideSubdirectory(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	sub := filepath.Join(dir, "pkg", "foo")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cdTo(t, sub)

	root, err := git.RepoRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(dir=%q): %v", dir, err)
	}
	got, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks(root=%q): %v", root, err)
	}
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestRepoRoot_OutsideRepo(t *testing.T) {
	requireGit(t)

	// Use a directory we control that is guaranteed not to be inside a git repo.
	outside := t.TempDir()
	cdTo(t, outside)

	_, err := git.RepoRoot()
	if err == nil {
		t.Fatal("expected an error outside a git repo, got nil")
	}
	if !strings.Contains(err.Error(), "not inside a git repository") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── ListWorktrees ─────────────────────────────────────────────────────────────

func TestListWorktrees_MainOnly(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	trees, err := git.ListWorktrees()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}

	wt := trees[0]
	if !wt.IsMain {
		t.Error("want IsMain=true for the only worktree")
	}
	if wt.Branch != "main" {
		t.Errorf("want branch 'main', got %q", wt.Branch)
	}
	if wt.HEAD == "" {
		t.Error("HEAD should not be empty")
	}
}

func TestListWorktrees_AfterAdd(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	wtPath := filepath.Join(t.TempDir(), "myrepo-second")
	if err := git.WorktreeAdd(wtPath, "second", true); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	t.Cleanup(func() { _ = git.WorktreeRemove(wtPath, true) })

	trees, err := git.ListWorktrees()
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	if len(trees) != 2 {
		t.Fatalf("want 2 worktrees after add, got %d", len(trees))
	}

	mainCount := 0
	for _, wt := range trees {
		if wt.IsMain {
			mainCount++
		}
	}
	if mainCount != 1 {
		t.Errorf("want exactly 1 IsMain worktree, got %d", mainCount)
	}
}

// ── WorktreeAdd / WorktreeRemove ──────────────────────────────────────────────

func TestWorktreeAdd_NewBranch(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	wtPath := filepath.Join(t.TempDir(), "myrepo-feat")
	if err := git.WorktreeAdd(wtPath, "feat/test", true); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	t.Cleanup(func() { _ = git.WorktreeRemove(wtPath, true) })

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree directory was not created at %s", wtPath)
	}

	trees, err := git.ListWorktrees()
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	found := false
	for _, wt := range trees {
		if wt.Branch == "feat/test" {
			found = true
		}
	}
	if !found {
		t.Error("new branch 'feat/test' not found in worktree list")
	}
}

func TestWorktreeAdd_ExistingBranch(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	runIn(t, dir, "git", "branch", "existing-branch")

	wtPath := filepath.Join(t.TempDir(), "myrepo-existing")
	if err := git.WorktreeAdd(wtPath, "existing-branch", false); err != nil {
		t.Fatalf("WorktreeAdd existing branch: %v", err)
	}
	t.Cleanup(func() { _ = git.WorktreeRemove(wtPath, true) })

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree directory not created at %s", wtPath)
	}
}

func TestWorktreeRemove_RemovesWorktree(t *testing.T) {
	requireGit(t)

	dir := initRepo(t, t.TempDir())
	cdTo(t, dir)

	wtPath := filepath.Join(t.TempDir(), "myrepo-to-remove")
	if err := git.WorktreeAdd(wtPath, "branch-to-remove", true); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	if err := git.WorktreeRemove(wtPath, false); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}

	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree directory still exists after removal")
	}

	trees, _ := git.ListWorktrees()
	for _, wt := range trees {
		if wt.Path == wtPath {
			t.Error("removed worktree still appears in ListWorktrees")
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// requireGit skips the test if git is not available in PATH.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
}

// initRepo creates a minimal git repository with one commit in dir.
// Requires git >= 2.28 (released July 2020).
func initRepo(t *testing.T, dir string) string {
	t.Helper()

	runIn(t, dir, "git", "init", "-b", "main")
	runIn(t, dir, "git", "config", "user.email", "test@example.com")
	runIn(t, dir, "git", "config", "user.name", "Test")

	writeFile(t, filepath.Join(dir, "README.md"), "# test")
	runIn(t, dir, "git", "add", ".")
	runIn(t, dir, "git", "commit", "-m", "init")

	return dir
}

// runIn runs a command inside dir, failing the test on any error.
func runIn(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	//nolint:gosec // test helper runs known commands (git) with controlled arguments
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cmd %s %v: %v\n%s", name, args, err, out)
	}
}

// cdTo changes the working directory for the duration of the test.
func cdTo(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeFileMode(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	//nolint:gosec // path is constructed from known test temp dirs + constant names
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Errorf("file %s\n  want: %q\n   got: %q", path, want, string(data))
	}
}
