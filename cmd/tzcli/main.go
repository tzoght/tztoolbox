// Command tzcli is the native CLI for the tztoolbox repository. It renders
// the canonical artifacts under shared/ and overrides/ into per-tool native
// trees (.cursor, .claude, .codex), installs those trees into the user's
// home, and provides authoring scaffolds, validation, and diagnostics.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/ui"
)

// Version is overridden at build time via -ldflags "-X main.Version=...".
var Version = "dev"

// Commit is overridden at build time via -ldflags "-X main.Commit=...".
var Commit = ""

// Date is overridden at build time via -ldflags "-X main.Date=...".
var Date = ""

func main() {
	root := newRootCmd()

	// Bare `tzcli` (no args) drops into the interactive TUI when stdout is
	// a terminal. Without a TTY (CI, pipes), we fall back to printing help
	// so the binary stays useful in non-interactive contexts.
	if len(os.Args) == 1 {
		if ui.StdoutIsTerminal() {
			if err := runTUI(root); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
		_ = root.Help()
		fmt.Fprintln(os.Stderr, "\n(no TTY detected; pass a subcommand or run from a terminal for the interactive shell)")
		return
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tzcli",
		Short: "Manage Cursor/Claude/Codex artifacts from one repo",
		Long: `tzcli is the native CLI for the tztoolbox repository.

It renders shared/ + overrides/ into the per-tool native trees
(.cursor/, .claude/, .codex/), installs those into your home,
and provides authoring scaffolds, validation, and diagnostics.`,
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       formatVersion(),
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			// Skip if the TUI host (or a test) already attached a renderer.
			if ui.HasRenderer(c.Context()) {
				return nil
			}
			mode := ui.Resolve(ui.ResolveOptions{
				NoColor:      flagNoColor,
				JSON:         flagJSON,
				NoColorEnv:   os.Getenv("NO_COLOR"),
				IsTerminalFn: ui.StdoutIsTerminal,
			})
			r := ui.NewRenderer(mode, c.OutOrStdout())
			c.SetContext(ui.With(c.Context(), r))
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&flagRepoRoot, "repo", "", "path to the tztoolbox repo (default: discover by walking up)")
	cmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable ANSI colors")
	cmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "emit machine-readable JSON where supported")

	cmd.AddCommand(
		newSyncCmd(),
		newInstallCmd(),
		newInstalledCmd(),
		newDoctorCmd(),
		newUpdateCmd(),
		newAddCmd(),
		newListCmd(),
		newValidateCmd(),
		newEnableCmd(),
		newDisableCmd(),
		newSearchCmd(),
		newShellCmd(),
	)
	return cmd
}

// Persistent flags shared across subcommands.
var (
	flagRepoRoot string
	flagVerbose  bool
	flagNoColor  bool
	flagJSON     bool
)

func formatVersion() string {
	if Commit == "" && Date == "" {
		return Version
	}
	return fmt.Sprintf("%s (commit %s, %s)", Version, Commit, Date)
}
