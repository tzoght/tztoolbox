package ui

import (
	"io"
	"os"

	"golang.org/x/term"
)

// Mode is the resolved output mode the CLI should use. Resolved once at
// startup and used both to construct a [Renderer] and to gate code that
// emits ANSI directly (like spinners).
type Mode int

const (
	// ModePlain prints uncoloured text via tabwriter. Used for pipes/CI.
	ModePlain Mode = iota
	// ModeStyled prints lipgloss-coloured text and bordered tables.
	ModeStyled
	// ModeJSON disables the renderer entirely; subcommands write JSON to
	// the underlying writer.
	ModeJSON
)

// String makes Mode log-friendly.
func (m Mode) String() string {
	switch m {
	case ModePlain:
		return "plain"
	case ModeStyled:
		return "styled"
	case ModeJSON:
		return "json"
	}
	return "unknown"
}

// ResolveOptions describes the inputs to [Resolve]. All fields are optional
// and have safe defaults; the CLI pulls them from persistent flags and env.
type ResolveOptions struct {
	NoColor       bool   // honors --no-color
	JSON          bool   // honors --json
	NoColorEnv    string // honors NO_COLOR (usually os.Getenv("NO_COLOR"))
	ForceTTY      bool   // override TTY detection (used by the TUI host)
	IsTerminalFn  func() bool
	StdoutFileFd  uintptr
	UseFdDetector bool // use term.IsTerminal(StdoutFileFd) instead of stdout sniffing
}

// Resolve computes the output [Mode] from opts. Precedence:
//
//  1. JSON wins (subcommand output is machine-readable).
//  2. NoColor / NO_COLOR / non-TTY -> plain.
//  3. Otherwise styled.
func Resolve(opts ResolveOptions) Mode {
	if opts.JSON {
		return ModeJSON
	}
	if opts.NoColor {
		return ModePlain
	}
	if opts.NoColorEnv != "" {
		return ModePlain
	}
	if opts.ForceTTY {
		return ModeStyled
	}
	if opts.IsTerminalFn != nil {
		if opts.IsTerminalFn() {
			return ModeStyled
		}
		return ModePlain
	}
	if opts.UseFdDetector {
		if term.IsTerminal(int(opts.StdoutFileFd)) {
			return ModeStyled
		}
		return ModePlain
	}
	// Conservative default: plain.
	return ModePlain
}

// NewRenderer constructs the right [Renderer] for the given mode. JSON mode
// is a no-op renderer that discards everything (subcommands write JSON to
// the underlying io.Writer directly).
func NewRenderer(mode Mode, w io.Writer) Renderer {
	switch mode {
	case ModeStyled:
		return NewLip(w, nil)
	case ModeJSON:
		return NewText(io.Discard)
	}
	return NewText(w)
}

// StdoutIsTerminal reports whether os.Stdout currently points at a terminal.
// Convenience wrapper used by main / TUI host.
func StdoutIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
