// Package config loads and writes the .wtconfig.yml configuration file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFile is the name of the wt configuration file placed in the repo root.
const ConfigFile = ".wtconfig.yml"

// DefaultCopyPaths are used when no .wtconfig.yml is found.
var DefaultCopyPaths = []string{
	".claude",
	".serena/project.yml",
	"CLAUDE.md",
}

// Config holds the parsed contents of .wtconfig.yml.
type Config struct {
	Copy []string `yaml:"copy"`
}

// Load reads .wtconfig.yml from repoRoot. Falls back to defaults if not found.
// Always returns a non-nil *Config when err is nil.
func Load(repoRoot string) (*Config, error) {
	cfgPath := filepath.Join(repoRoot, ConfigFile)

	//nolint:gosec // path is constructed from repo root (git-trusted) + constant
	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return &Config{Copy: DefaultCopyPaths}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Copy) == 0 {
		cfg.Copy = DefaultCopyPaths
	}

	return &cfg, nil
}

// WriteDefault creates a .wtconfig.yml with sensible defaults in dir.
func WriteDefault(dir string) error {
	content := `# wt configuration
# List of files and directories to copy into each new worktree.
# Paths are relative to the repository root.
# Missing paths are silently skipped.
copy:
  - .claude/
  - .serena/project.yml
  - CLAUDE.md
  # - .cursor/rules       # uncomment if you use Cursor
  # - .env.example
`
	return os.WriteFile(filepath.Join(dir, ConfigFile), []byte(content), 0o600)
}
