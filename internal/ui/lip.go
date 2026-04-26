package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// lipRenderer styles output with lipgloss for an interactive TTY. Tables
// have a rounded border with subtle column dividers; cells inherit colors
// from their [Style]. Unicode glyphs are used where they help legibility.
type lipRenderer struct {
	w  io.Writer
	st *Theme
}

// Theme bundles the lipgloss styles used by the lip renderer. Exported so
// the TUI host can reuse the same color tokens for chrome.
type Theme struct {
	Heading lipgloss.Style
	Note    lipgloss.Style
	Warn    lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
	Muted   lipgloss.Style
	Bold    lipgloss.Style
	Path    lipgloss.Style
	Tool    lipgloss.Style
	Header  lipgloss.Style // table header row
	Border  lipgloss.Style // rounded border around tables
}

// DefaultTheme returns the default lipgloss color palette. It intentionally
// uses adaptive colors so it looks correct on both dark and light terminals.
func DefaultTheme() *Theme {
	return &Theme{
		Heading: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#1f3a5f", Dark: "#7dd3fc"}),
		Note:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#374151", Dark: "#cbd5e1"}),
		Warn:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#fbbf24"}),
		Error:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f87171"}),
		Success: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#34d399"}),
		Muted:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}),
		Bold:    lipgloss.NewStyle().Bold(true),
		Path:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0369a1", Dark: "#7dd3fc"}),
		Tool:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#c4b5fd"}),
		Header:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#0f172a", Dark: "#f8fafc"}),
		Border:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#475569"}),
	}
}

// NewLip returns a styled [Renderer] writing to w. theme may be nil to use
// [DefaultTheme]. Used when stdout is a TTY and colors are enabled.
func NewLip(w io.Writer, theme *Theme) Renderer {
	if w == nil {
		w = io.Discard
	}
	if theme == nil {
		theme = DefaultTheme()
	}
	return &lipRenderer{w: w, st: theme}
}

// styleFor returns the lipgloss.Style for a [Style] token.
func (r *lipRenderer) styleFor(s Style) lipgloss.Style {
	switch s {
	case StyleSuccess:
		return r.st.Success
	case StyleWarn:
		return r.st.Warn
	case StyleError:
		return r.st.Error
	case StyleMuted:
		return r.st.Muted
	case StyleBold:
		return r.st.Bold
	case StylePath:
		return r.st.Path
	case StyleTool:
		return r.st.Tool
	case StyleHeading:
		return r.st.Heading
	}
	return lipgloss.NewStyle()
}

func (r *lipRenderer) Heading(s string) {
	fmt.Fprintln(r.w, r.st.Heading.Render(s))
}

func (r *lipRenderer) Note(s string) {
	fmt.Fprintln(r.w, r.st.Note.Render(s))
}

func (r *lipRenderer) Warn(s string) {
	fmt.Fprintln(r.w, r.st.Warn.Render("warn: "+s))
}

func (r *lipRenderer) Errorf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	msg = strings.TrimRight(msg, "\n")
	fmt.Fprintln(r.w, r.st.Error.Render("error: "+msg))
}

// Table renders a bordered table. We compute column widths from the widest
// (header or cell) value, then pad each cell with lipgloss before joining
// rows with column separators. The border is drawn with rounded box-drawing.
func (r *lipRenderer) Table(headers []string, rows [][]Cell) {
	if len(headers) == 0 && len(rows) == 0 {
		return
	}

	cols := len(headers)
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}

	widths := make([]int, cols)
	for i, h := range headers {
		if w := lipgloss.Width(h); w > widths[i] {
			widths[i] = w
		}
	}
	for _, row := range rows {
		for i, c := range row {
			if w := lipgloss.Width(c.Text); w > widths[i] {
				widths[i] = w
			}
		}
	}

	rowToString := func(headerRow bool, cells []string, styles []Style) string {
		parts := make([]string, len(widths))
		for i := range widths {
			text := ""
			st := StyleDefault
			if i < len(cells) {
				text = cells[i]
			}
			if i < len(styles) {
				st = styles[i]
			}
			padded := lipgloss.NewStyle().Width(widths[i]).Render(text)
			if headerRow {
				padded = r.st.Header.Width(widths[i]).Render(text)
			} else if st != StyleDefault {
				padded = r.styleFor(st).Width(widths[i]).Render(text)
			}
			parts[i] = padded
		}
		sep := r.st.Border.Render(" | ")
		return r.st.Border.Render("| ") + strings.Join(parts, sep) + r.st.Border.Render(" |")
	}

	var sb strings.Builder
	// Top border
	sb.WriteString(r.borderLine(widths, "┌", "┬", "┐"))
	sb.WriteByte('\n')

	if len(headers) > 0 {
		styles := make([]Style, len(headers))
		sb.WriteString(rowToString(true, headers, styles))
		sb.WriteByte('\n')
		sb.WriteString(r.borderLine(widths, "├", "┼", "┤"))
		sb.WriteByte('\n')
	}

	for _, row := range rows {
		texts := make([]string, len(row))
		styles := make([]Style, len(row))
		for i, c := range row {
			texts[i] = c.Text
			styles[i] = c.Style
		}
		sb.WriteString(rowToString(false, texts, styles))
		sb.WriteByte('\n')
	}
	sb.WriteString(r.borderLine(widths, "└", "┴", "┘"))
	sb.WriteByte('\n')

	fmt.Fprint(r.w, sb.String())
}

// borderLine builds a horizontal box-drawing border with the given junction
// glyphs. widths includes the cell content width only; +2 accounts for the
// inner cell padding (" cell ").
func (r *lipRenderer) borderLine(widths []int, left, mid, right string) string {
	var sb strings.Builder
	sb.WriteString(left)
	for i, w := range widths {
		sb.WriteString(strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			sb.WriteString(mid)
		}
	}
	sb.WriteString(right)
	return r.st.Border.Render(sb.String())
}

func (r *lipRenderer) KeyValues(pairs ...KV) {
	if len(pairs) == 0 {
		return
	}
	keyW := 0
	for _, p := range pairs {
		if w := lipgloss.Width(p.Key); w > keyW {
			keyW = w
		}
	}
	for _, p := range pairs {
		key := r.st.Muted.Width(keyW).Render(p.Key)
		val := p.Value
		if p.Style != StyleDefault {
			val = r.styleFor(p.Style).Render(p.Value)
		}
		fmt.Fprintf(r.w, "%s  %s\n", key, val)
	}
}

func (r *lipRenderer) Raw(s string) {
	fmt.Fprint(r.w, s)
}

func (r *lipRenderer) Flush() {}
