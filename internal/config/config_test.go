package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonioliveira/wt/internal/config"
)

// writeFile is a test helper that creates a file with the given content,
// failing the test immediately on any error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestLoad_NoConfigFile verifies that a missing .wtconfig.yml returns the
// default copy paths without an error.
func TestLoad_NoConfigFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	assertEqualPaths(t, config.DefaultCopyPaths, cfg.Copy)
}

// TestLoad_ValidConfig verifies that a well-formed .wtconfig.yml is parsed
// correctly and its copy paths are returned as-is.
func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), `
copy:
  - .claude/
  - .serena/project.yml
  - .env.example
`)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{".claude/", ".serena/project.yml", ".env.example"}
	assertEqualPaths(t, want, cfg.Copy)
}

// TestLoad_EmptyCopyList verifies that an empty copy list in the config file
// falls back to the defaults rather than returning an empty slice.
func TestLoad_EmptyCopyList(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), `copy: []`)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqualPaths(t, config.DefaultCopyPaths, cfg.Copy)
}

// TestLoad_EmptyFile verifies that a completely empty config file falls back
// to the defaults.
func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), "")

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqualPaths(t, config.DefaultCopyPaths, cfg.Copy)
}

// TestLoad_CommentsOnly verifies that a config file containing only comments
// (no keys) falls back to defaults.
func TestLoad_CommentsOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), `
# wt configuration
# copy:
#   - .claude/
`)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqualPaths(t, config.DefaultCopyPaths, cfg.Copy)
}

// TestLoad_InvalidYAML verifies that a malformed config file returns an error.
func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), `copy: [unclosed`)

	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("expected an error for invalid YAML, got nil")
	}
}

// TestLoad_SingleEntry verifies that a single-entry copy list is handled
// correctly (no off-by-one or slice issues).
func TestLoad_SingleEntry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".wtconfig.yml"), `
copy:
  - CLAUDE.md
`)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqualPaths(t, []string{"CLAUDE.md"}, cfg.Copy)
}

// TestLoad_NeverReturnsNil verifies the contract that Load always returns
// a non-nil *Config when err is nil.
func TestLoad_NeverReturnsNil(t *testing.T) {
	cases := []struct {
		name    string
		content *string // nil means no file
	}{
		{"no file", nil},
		{"empty file", strPtr("")},
		{"valid file", strPtr("copy:\n  - .claude/\n")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.content != nil {
				writeFile(t, filepath.Join(dir, ".wtconfig.yml"), *tc.content)
			}

			cfg, err := config.Load(dir)
			if err != nil {
				t.Skipf("skipping nil check — error returned: %v", err)
			}
			if cfg == nil {
				t.Error("Load returned nil config with nil error")
			}
		})
	}
}

// TestWriteDefault verifies that WriteDefault creates a parseable .wtconfig.yml
// that produces non-empty copy paths when loaded back.
func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()

	if err := config.WriteDefault(dir); err != nil {
		t.Fatalf("WriteDefault error: %v", err)
	}

	// File must exist.
	cfgPath := filepath.Join(dir, ".wtconfig.yml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatalf(".wtconfig.yml was not created")
	}

	// File must be loadable and produce a non-empty copy list.
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load after WriteDefault error: %v", err)
	}
	if len(cfg.Copy) == 0 {
		t.Error("WriteDefault produced a config with no copy paths")
	}
}

// TestWriteDefault_Idempotent verifies that calling WriteDefault twice does
// not corrupt the file (second call overwrites cleanly).
func TestWriteDefault_Idempotent(t *testing.T) {
	dir := t.TempDir()

	for i := range 2 {
		if err := config.WriteDefault(dir); err != nil {
			t.Fatalf("WriteDefault call %d error: %v", i+1, err)
		}
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load after double WriteDefault: %v", err)
	}
	if len(cfg.Copy) == 0 {
		t.Error("config has no copy paths after double WriteDefault")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func assertEqualPaths(t *testing.T, want, got []string) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("path count mismatch\n  want %d: %v\n   got %d: %v",
			len(want), want, len(got), got)
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("path[%d] mismatch\n  want: %q\n   got: %q", i, want[i], got[i])
		}
	}
}

func strPtr(s string) *string { return &s }
