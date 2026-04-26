package tui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"mvdan.cc/sh/v3/shell"

	"github.com/tzoght/tztoolbox/internal/ui"
)

// RootCmdFactory is the contract the host uses to give the TUI a Cobra
// command tree. Each REPL invocation calls it to obtain a *fresh* tree so
// flags don't bleed across runs.
type RootCmdFactory func() *cobra.Command

// commandDoneMsg is posted when a single REPL line has finished executing.
// Output already contains the styled body the renderer wrote.
type commandDoneMsg struct {
	line     string
	head     string // canonical first token after mnemonic expansion (e.g. "add")
	output   string
	err      error
	duration time.Duration
}

// runLine executes a single REPL line in a tea.Cmd. It splits the line into
// argv via POSIX shell quoting, then runs a fresh Cobra tree with all
// output funneled through a styled ui.Renderer pointed at a buffer.
func (m Model) runLine(line string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		// Expand REPL-only mnemonics (e.g. "s" -> "sync", "?add" stays as
		// "?" since "?" is mapped to "help").
		expanded := expandMnemonic(line)

		argv, err := shell.Fields(expanded, nil)
		if err != nil {
			return commandDoneMsg{
				line:     line,
				err:      fmt.Errorf("parse error: %w", err),
				duration: time.Since(start),
			}
		}
		argv = stripBareName(argv)
		if len(argv) == 0 {
			return commandDoneMsg{line: line, duration: time.Since(start)}
		}
		head := argv[0]
		// Built-ins handled inside the model (quit/help/clear) shouldn't
		// reach this code path; treat them defensively as no-ops.
		switch argv[0] {
		case "quit", "exit", ":q", ":quit":
			return tea.Quit()
		}

		root := m.factory()
		if root == nil {
			return commandDoneMsg{
				line:     line,
				err:      fmt.Errorf("no Cobra root command available"),
				duration: time.Since(start),
			}
		}

		var buf bytes.Buffer
		root.SetOut(&buf)
		root.SetErr(&buf)
		root.SetArgs(argv)
		// We render errors ourselves (with a help hint) and prefer the
		// command's full usage block when an arg-validator fails — that
		// turns "accepts 2 arg(s), received 0" into something actionable.
		root.SilenceUsage = false
		root.SilenceErrors = true

		// Attach a styled renderer so the existing PersistentPreRunE in
		// main.go skips its auto-detect path.
		mode := ui.ModeStyled
		if m.flagJSONFromArgs(argv) {
			mode = ui.ModeJSON
		}
		r := ui.NewRenderer(mode, &buf)
		ctx := ui.With(context.Background(), r)
		root.SetContext(ctx)

		execErr := root.Execute()
		r.Flush()

		return commandDoneMsg{
			line:     line,
			head:     head,
			output:   buf.String(),
			err:      execErr,
			duration: time.Since(start),
		}
	}
}

// stripBareName drops a leading "tzcli" token if the user typed it. We
// accept both "tzcli sync" and "sync" inside the REPL.
func stripBareName(argv []string) []string {
	if len(argv) > 0 && argv[0] == "tzcli" {
		return argv[1:]
	}
	return argv
}

// flagJSONFromArgs reports whether the user passed --json on the command
// line. Used to switch the REPL renderer to discard mode so JSON is emitted
// raw to the buffer (matches one-shot behavior).
func (m Model) flagJSONFromArgs(argv []string) bool {
	for _, a := range argv {
		if a == "--json" || a == "-json" || strings.HasPrefix(a, "--json=") {
			return true
		}
	}
	return false
}
