package git

import (
	"strings"
	"testing"
)

func TestParseWorktreeList_Single(t *testing.T) {
	raw := `worktree /home/user/myrepo
HEAD abc1234567890abcdef
branch refs/heads/main

`
	trees := parseWorktreeList(raw)

	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}

	wt := trees[0]
	assertEqual(t, "/home/user/myrepo", wt.Path)
	assertEqual(t, "main", wt.Branch)
	assertEqual(t, "abc12345", wt.HEAD) // truncated to 8 chars
	if !wt.IsMain {
		t.Error("first worktree should be marked as main")
	}
}

func TestParseWorktreeList_Multiple(t *testing.T) {
	raw := `worktree /home/user/myrepo
HEAD aaaaaaaaaaaaaaaa
branch refs/heads/main

worktree /home/user/myrepo-feature
HEAD bbbbbbbbbbbbbbbb
branch refs/heads/feature/my-feature

`
	trees := parseWorktreeList(raw)

	if len(trees) != 2 {
		t.Fatalf("want 2 worktrees, got %d", len(trees))
	}

	if !trees[0].IsMain {
		t.Error("first worktree should be IsMain=true")
	}
	if trees[1].IsMain {
		t.Error("second worktree should be IsMain=false")
	}

	assertEqual(t, "main", trees[0].Branch)
	assertEqual(t, "feature/my-feature", trees[1].Branch)
	assertEqual(t, "/home/user/myrepo-feature", trees[1].Path)
}

func TestParseWorktreeList_DetachedHEAD(t *testing.T) {
	raw := `worktree /home/user/myrepo
HEAD abc1234567890abcdef
detached

`
	trees := parseWorktreeList(raw)

	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}
	assertEqual(t, "(detached)", trees[0].Branch)
}

func TestParseWorktreeList_HEADTruncatedTo8(t *testing.T) {
	raw := `worktree /home/user/myrepo
HEAD 0123456789abcdef0123456789abcdef
branch refs/heads/main

`
	trees := parseWorktreeList(raw)

	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}
	assertEqual(t, "01234567", trees[0].HEAD)
}

func TestParseWorktreeList_HEADShortHashNotTruncated(t *testing.T) {
	// A HEAD hash shorter than 8 chars should be kept as-is.
	raw := `worktree /home/user/myrepo
HEAD abc123
branch refs/heads/main

`
	trees := parseWorktreeList(raw)

	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}
	assertEqual(t, "abc123", trees[0].HEAD)
}

func TestParseWorktreeList_StripsBranchPrefix(t *testing.T) {
	// Ensure refs/heads/ is stripped from branch names.
	raw := `worktree /home/user/myrepo
HEAD aaaaaaaaaaaaaaaa
branch refs/heads/feat/some-long/branch-name

`
	trees := parseWorktreeList(raw)

	assertEqual(t, "feat/some-long/branch-name", trees[0].Branch)
}

func TestParseWorktreeList_Empty(t *testing.T) {
	trees := parseWorktreeList("")
	if len(trees) != 0 {
		t.Errorf("want 0 worktrees for empty input, got %d", len(trees))
	}
}

func TestParseWorktreeList_NoTrailingNewline(t *testing.T) {
	// git output sometimes omits the trailing blank line — parser must not drop
	// the last worktree.
	raw := `worktree /home/user/myrepo
HEAD abc1234567890abcdef
branch refs/heads/main`

	trees := parseWorktreeList(raw)

	if len(trees) != 1 {
		t.Fatalf("want 1 worktree, got %d", len(trees))
	}
	assertEqual(t, "main", trees[0].Branch)
}

func TestParseWorktreeList_ThreeWorktrees(t *testing.T) {
	raw := `worktree /home/user/myrepo
HEAD aaaaaaaaaaaaaaaa
branch refs/heads/main

worktree /home/user/myrepo-feat-a
HEAD bbbbbbbbbbbbbbbb
branch refs/heads/feat/a

worktree /home/user/myrepo-feat-b
HEAD cccccccccccccccc
branch refs/heads/feat/b

`
	trees := parseWorktreeList(raw)

	if len(trees) != 3 {
		t.Fatalf("want 3 worktrees, got %d", len(trees))
	}

	assertEqual(t, "main", trees[0].Branch)
	assertEqual(t, "feat/a", trees[1].Branch)
	assertEqual(t, "feat/b", trees[2].Branch)

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

// ── helpers ──────────────────────────────────────────────────────────────────

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("\n  want: %q\n   got: %q", want, got)
	}
}

// ── parseGitVersion ───────────────────────────────────────────────────────────

func TestParseGitVersion(t *testing.T) {
	cases := []struct {
		input        string
		wantMajor    int
		wantMinor    int
		wantErrMatch string
	}{
		{"git version 2.39.1", 2, 39, ""},
		{"git version 2.28.0", 2, 28, ""},
		{"git version 2.27.0", 2, 27, ""},
		{"git version 2.39.1.windows.1", 2, 39, ""}, // Windows suffix
		{"git version 1.8.3.1", 1, 8, ""},           // old format
		{"git version 2.39", 2, 39, ""},             // no patch
		{"not a version", 0, 0, "unexpected format"},
		{"git version abc.def.0", 0, 0, "parse major"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			major, minor, err := parseGitVersion(tc.input)

			if tc.wantErrMatch != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil", tc.wantErrMatch)
				}
				if !strings.Contains(err.Error(), tc.wantErrMatch) {
					t.Errorf("want error %q, got %q", tc.wantErrMatch, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if major != tc.wantMajor || minor != tc.wantMinor {
				t.Errorf("want %d.%d, got %d.%d", tc.wantMajor, tc.wantMinor, major, minor)
			}
		})
	}
}
