// Package ui defines the rendering abstraction every tzcli subcommand uses
// to emit output. Two concrete implementations exist:
//
//   - text.Renderer: plain ASCII via text/tabwriter for piped/CI output.
//   - lip.Renderer:  lipgloss-styled with colors and bordered tables for TTY.
//
// The right renderer is chosen at startup by [Resolve] from a [Mode] value
// derived from --no-color, NO_COLOR, --json, isatty(stdout), and the TUI
// host's explicit override. Subcommands access it through [From] after main
// has called [With].
package ui

import (
	"context"
	"io"
)

// Style classifies the visual treatment of a [Cell] or single output line.
// Concrete renderers map these to ANSI codes (lip) or ignore them (text).
type Style int

const (
	StyleDefault Style = iota
	StyleSuccess       // ok / in_sync
	StyleWarn          // changed / drift / orphan
	StyleError         // hard error
	StyleMuted         // dashes, dim placeholders
	StyleBold          // emphasis without color shift
	StylePath          // file/dir paths
	StyleTool          // tool names (cursor, claude, codex)
	StyleHeading       // section heading
)

// Cell is a single table cell with optional styling.
type Cell struct {
	Text  string
	Style Style
}

// PlainCell is sugar for an unstyled [Cell].
func PlainCell(text string) Cell { return Cell{Text: text, Style: StyleDefault} }

// StyledCell is sugar for a styled [Cell].
func StyledCell(text string, style Style) Cell { return Cell{Text: text, Style: style} }

// KV is a single key/value pair printed by [Renderer.KeyValues].
type KV struct {
	Key   string
	Value string
	Style Style // applied to value
}

// Renderer is the abstract sink every subcommand writes through.
//
// Implementations must be safe to use after [Renderer.Flush] returns; CLI
// invocations call Flush at most once at program exit but the TUI may flush
// after every command.
type Renderer interface {
	Heading(s string)
	Note(s string)
	Warn(s string)
	Errorf(format string, a ...any)
	Table(headers []string, rows [][]Cell)
	KeyValues(pairs ...KV)
	// Raw writes pre-formatted text verbatim (used for unified diffs etc).
	Raw(s string)
	Flush()
}

// rendererKey is the unexported context key used by [With] / [From].
type rendererKey struct{}

// With returns ctx with r attached so subcommands can call [From].
func With(ctx context.Context, r Renderer) context.Context {
	return context.WithValue(ctx, rendererKey{}, r)
}

// From returns the renderer attached to ctx, or a no-op text renderer
// targeting [io.Discard] if none was set. Subcommands should always be able
// to call this without a nil-check.
func From(ctx context.Context) Renderer {
	if ctx == nil {
		return NewText(io.Discard)
	}
	if r, ok := ctx.Value(rendererKey{}).(Renderer); ok && r != nil {
		return r
	}
	return NewText(io.Discard)
}

// HasRenderer reports whether a Renderer has been attached to ctx via [With].
// Used by main's PersistentPreRunE to avoid clobbering a renderer that the
// TUI host already injected.
func HasRenderer(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	_, ok := ctx.Value(rendererKey{}).(Renderer)
	return ok
}
