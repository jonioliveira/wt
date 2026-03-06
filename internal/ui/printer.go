// Package ui provides shared terminal output helpers for the wt CLI.
package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Terminal color/style presets used across all commands.
var (
	// Bold is used for section headers and important values.
	Bold = color.New(color.Bold)
	// Green indicates a successful operation.
	Green = color.New(color.FgGreen)
	// Yellow indicates a warning or partial failure.
	Yellow = color.New(color.FgYellow)
	// Cyan is used for secondary information such as branch names.
	Cyan = color.New(color.FgCyan)
	// Dim is used for de-emphasized output such as skipped paths.
	Dim = color.New(color.Faint)
)

// SyncResult summarises the outcome of a context file sync operation.
type SyncResult struct {
	Copied  int
	Skipped int
	Failed  int
}

// PrintSummary writes a one-line Done summary to w.
func (r SyncResult) PrintSummary(w io.Writer) {
	_, _ = Bold.Fprintf(w, "Done! %d copied", r.Copied)
	if r.Skipped > 0 {
		_, _ = fmt.Fprintf(w, ", %d skipped", r.Skipped)
	}
	if r.Failed > 0 {
		_, _ = Yellow.Fprintf(w, ", %d failed", r.Failed)
	}
	_, _ = fmt.Fprintln(w)
}
