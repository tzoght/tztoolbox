// Package tui implements the resident interactive shell mode for tzcli. It
// is invoked by `tzcli` (no arguments) on a TTY, or explicitly via
// `tzcli shell`. The host injects a [RootCmdFactory] that returns a fresh
// Cobra tree per dispatched line so flag state never bleeds between runs.
package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the TUI program, blocking until the user quits. It returns
// nil on a clean exit (Ctrl-D / quit), or an error if Bubbletea couldn't
// start (typically: not a TTY).
func Run(ctx context.Context, opts Options) error {
	if opts.Factory == nil {
		return fmt.Errorf("tui.Run: nil factory")
	}
	m := NewModel(opts)
	p := tea.NewProgram(m,
		tea.WithContext(ctx),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
