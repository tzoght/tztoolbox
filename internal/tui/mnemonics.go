package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// mnemonic binds a single-character shortcut to a tzcli subcommand and a
// human-readable description. Mnemonics are *only* expanded inside the
// REPL, so the standalone CLI surface (and shell scripts that wrap it)
// stay free of magic single-letter aliases.
type mnemonic struct {
	Key  string // e.g. "s" or "?"
	Cmd  string // canonical subcommand, e.g. "sync"
	Desc string
}

// mnemonics is the curated startup menu, ordered for display. The leftmost
// column shows the mnemonic; the next column the canonical command name.
//
// When picking keys, prefer the first letter of the canonical command and
// fall back to a memorable alternative on collision (e.g. installed=I,
// disable=x).
var mnemonics = []mnemonic{
	{"s", "sync", "render shared/+overrides/ → in-repo native trees"},
	{"i", "install", "copy in-repo trees → ~/.cursor, ~/.claude, ~/.codex"},
	{"I", "installed", "list deployed artifacts (status: in_sync/changed/...)"},
	{"d", "doctor", "environment readiness, detected tools, drift status"},
	{"u", "update", "fast-forward git pull from origin (refuses on dirty tree)"},
	{"a", "add", "scaffold a new artifact (command|rule|skill)"},
	{"l", "list", "list artifacts under shared/"},
	{"v", "validate", "lint the artifact corpus under shared/"},
	{"e", "enable", "re-enable an artifact for one tool"},
	{"x", "disable", "drop an artifact from one tool's render"},
	{"/", "search", "regex search across shared/ content"},
	{"?", "help", "show usage for a subcommand: 'help <name>' or '<name> --help'"},
	{"m", "menu", "show this menu again (also: F1, ctrl+g)"},
	{"c", "clear", "clear the output viewport"},
	{"q", "quit", "exit the shell (also: ctrl+d)"},
}

// expandMnemonic returns the canonical command line if line is a single
// mnemonic letter (with optional trailing args). Returns line unchanged if
// no expansion applies. Examples:
//
//	"s"          -> "sync"
//	"s --check"  -> "sync --check"
//	"sync"       -> "sync" (no change; already canonical)
//	""           -> "" (no change)
func expandMnemonic(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return line
	}
	// Find the first whitespace boundary to isolate the head token.
	head := trimmed
	rest := ""
	if i := strings.IndexAny(trimmed, " \t"); i >= 0 {
		head = trimmed[:i]
		rest = trimmed[i:]
	}
	for _, mn := range mnemonics {
		if mn.Key == head {
			return mn.Cmd + rest
		}
	}
	return line
}

// renderStartupMenu builds the colorized startup menu shown once when the
// REPL launches. The output is plain string with embedded ANSI styles; the
// caller is responsible for appending it to the viewport.
func renderStartupMenu(chrome *chromeStyles, version string) string {
	keyStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#c4b5fd"})
	cmdStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#1f3a5f", Dark: "#7dd3fc"})
	descStyle := chrome.Footer.Padding(0)
	titleStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#0f172a", Dark: "#f8fafc"})

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(fmt.Sprintf("welcome to tzcli %s", version)))
	sb.WriteString("\n")
	sb.WriteString(descStyle.Render("type a command name, a mnemonic letter, or 'help <name>' for usage."))
	sb.WriteString("\n\n")

	// Compute column widths from the live mnemonic list so additions don't
	// drift the layout.
	keyW, cmdW := 0, 0
	for _, mn := range mnemonics {
		if l := len(mn.Key); l > keyW {
			keyW = l
		}
		if l := len(mn.Cmd); l > cmdW {
			cmdW = l
		}
	}

	for _, mn := range mnemonics {
		key := keyStyle.Render(fmt.Sprintf("[%s]", padRight(mn.Key, keyW)))
		cmd := cmdStyle.Render(padRight(mn.Cmd, cmdW))
		desc := descStyle.Render(mn.Desc)
		fmt.Fprintf(&sb, "  %s  %s   %s\n", key, cmd, desc)
	}

	sb.WriteString("\n")
	sb.WriteString(descStyle.Render(
		"shortcuts: F1/ctrl+g=this menu  tab=complete  ctrl+k=palette  ctrl+l=clear  ?=key help  ctrl+d=quit",
	))
	sb.WriteString("\n")
	return sb.String()
}

// padRight pads s on the right with spaces so its rune count equals width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// mnemonicPaletteItems exposes mnemonics for the command palette so users
// can fuzzy-match by either letter or full name.
func mnemonicPaletteItems() []paletteItem {
	out := make([]paletteItem, 0, len(mnemonics))
	for _, mn := range mnemonics {
		out = append(out, paletteItem{
			command: mn.Cmd,
			desc:    fmt.Sprintf("[%s]  %s", mn.Key, mn.Desc),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].command < out[j].command })
	return out
}
