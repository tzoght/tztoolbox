package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// chromeStyles bundles lipgloss styles used by the TUI shell (header bar,
// prompt chrome, footer hints). Distinct from the shared ui.Theme used by
// rendered command output: chrome must contrast against the body to give
// the user a clear "where am I" cue.
type chromeStyles struct {
	Header    lipgloss.Style
	HeaderDim lipgloss.Style
	Footer    lipgloss.Style
	Prompt    lipgloss.Style
	PromptCmd lipgloss.Style
	Spinner   lipgloss.Style
	Palette   lipgloss.Style
	Border    lipgloss.Style
	Time      lipgloss.Style
	OK        lipgloss.Style
	Warn      lipgloss.Style
	Err       lipgloss.Style
}

func defaultChromeStyles() *chromeStyles {
	return &chromeStyles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#0f172a", Dark: "#f8fafc"}).
			Background(lipgloss.AdaptiveColor{Light: "#dbeafe", Dark: "#1e3a8a"}).
			Padding(0, 1),
		HeaderDim: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#475569", Dark: "#94a3b8"}).
			Background(lipgloss.AdaptiveColor{Light: "#dbeafe", Dark: "#1e3a8a"}).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#475569", Dark: "#94a3b8"}).
			Padding(0, 1),
		Prompt: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#c4b5fd"}),
		PromptCmd: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#1f3a5f", Dark: "#7dd3fc"}),
		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#c4b5fd"}),
		Palette: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#475569"}).
			Padding(0, 1),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#475569"}),
		Time: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#64748b"}),
		OK:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#34d399"}),
		Warn: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#fbbf24"}),
		Err:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f87171"}),
	}
}
