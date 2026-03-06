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

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a .wtconfig.yml with default settings",
		Long: `Creates a .wtconfig.yml in the repository root with the default list of
context files to copy into new worktrees. Edit it to add or remove paths.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				return err
			}

			cfgPath := filepath.Join(repoRoot, config.ConfigFile)

			if _, err := os.Stat(cfgPath); err == nil {
				_, _ = ui.Yellow.Printf(".wtconfig.yml already exists at %s\n", cfgPath)
				return nil
			}

			if err := config.WriteDefault(repoRoot); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			_, _ = ui.Green.Printf("✓ Created %s\n\n", cfgPath)
			fmt.Println("Edit it to configure which files are synced into new worktrees.")
			fmt.Println("Commit it to share the config with your team," +
				" or add it to .gitignore to keep it personal.")

			return nil
		},
	}
}
