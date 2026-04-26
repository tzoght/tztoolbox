package ui

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// textRenderer is the plain renderer: no ANSI, predictable for pipes/CI.
//
// It buffers table writes through a tabwriter so columns stay aligned.
// All other primitives go straight to the underlying writer.
type textRenderer struct {
	w io.Writer
}

// NewText returns a [Renderer] that writes plain text without ANSI to w.
// Used in CI, when stdout is piped, or when --no-color / NO_COLOR is set.
func NewText(w io.Writer) Renderer {
	if w == nil {
		w = io.Discard
	}
	return &textRenderer{w: w}
}

func (r *textRenderer) Heading(s string) {
	fmt.Fprintln(r.w, s)
}

func (r *textRenderer) Note(s string) {
	fmt.Fprintln(r.w, s)
}

func (r *textRenderer) Warn(s string) {
	fmt.Fprintln(r.w, s)
}

func (r *textRenderer) Errorf(format string, a ...any) {
	fmt.Fprintf(r.w, format, a...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Fprintln(r.w)
	}
}

func (r *textRenderer) Table(headers []string, rows [][]Cell) {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	if len(headers) > 0 {
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		parts := make([]string, len(row))
		for i, c := range row {
			parts[i] = c.Text
		}
		fmt.Fprintln(tw, strings.Join(parts, "\t"))
	}
	_ = tw.Flush()
}

func (r *textRenderer) KeyValues(pairs ...KV) {
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	for _, p := range pairs {
		fmt.Fprintf(tw, "%s:\t%s\n", p.Key, p.Value)
	}
	_ = tw.Flush()
}

func (r *textRenderer) Raw(s string) {
	fmt.Fprint(r.w, s)
}

func (r *textRenderer) Flush() {}
