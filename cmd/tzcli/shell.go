package main

import (
	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/tui"
)

// newShellCmd registers the explicit `tzcli shell` subcommand. It is
// equivalent to running bare `tzcli` on a TTY (which auto-routes here).
func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Launch the interactive tzcli TUI",
		Long: `Open a resident, Bubbletea-based shell. Subsequent commands
(sync, install, doctor, ...) run inside that shell with colored output and
formatted tables. Press ctrl+k for the command palette, ? for help, ctrl+d
to quit.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTUI(cmd)
		},
	}
}

// runTUI is the shared entry point used both by `tzcli shell` and by the
// bare-invocation route in main.go. It resolves the repo root (best-effort)
// and hands the model a factory that always returns a fresh root command.
func runTUI(cmd *cobra.Command) error {
	root, _ := resolveRoot() // best-effort; TUI works without a repo
	return tui.Run(cmd.Context(), tui.Options{
		Factory:  newRootCmd,
		RepoRoot: root,
		Version:  formatVersion(),
	})
}
