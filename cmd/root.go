// Package cmd implements the wt command-line interface.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jonioliveira/wt/internal/git"
)

// NewRootCmd builds and returns the root cobra command with all subcommands
// registered. Callers own the lifecycle; there is no package-level state.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "wt",
		Short: "Git worktrees with AI context — create, sync, and remove worktrees",
		Long: `wt is a thin wrapper around git worktree that automatically copies
your AI assistant context files (.claude/, .serena/, CLAUDE.md) into
every new worktree so you can work in parallel without losing context.

Configure which files are copied by adding a .wtconfig.yml to your repo root.`,

		// Validate git availability and version before any subcommand runs.
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return git.CheckVersion()
		},

		// Don't print usage on every error — it's noisy and rarely helpful.
		SilenceUsage: true,
		// Let main handle error printing so we control the format.
		SilenceErrors: true,
	}

	root.AddCommand(
		newNewCmd(),
		newRemoveCmd(),
		newListCmd(),
		newSyncCmd(),
		newInitCmd(),
	)

	return root
}
