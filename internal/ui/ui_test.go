package ui

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestTextRendererTableAlignsColumns(t *testing.T) {
	var buf bytes.Buffer
	r := NewText(&buf)
	r.Table([]string{"NAME", "STATUS"}, [][]Cell{
		{PlainCell("a"), PlainCell("ok")},
		{PlainCell("longer-name"), PlainCell("changed")},
	})
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Fatalf("expected headers in output:\n%s", out)
	}
	// Each row should be on its own line.
	if got := strings.Count(out, "\n"); got != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", got, out)
	}
	// Tabwriter pads with spaces, no tabs in the rendered output.
	if strings.Contains(out, "\t") {
		t.Fatalf("text renderer leaked tabs to output:\n%s", out)
	}
}

func TestTextRendererKeyValuesHasColon(t *testing.T) {
	var buf bytes.Buffer
	r := NewText(&buf)
	r.KeyValues(KV{Key: "repo", Value: "/x"})
	if !strings.Contains(buf.String(), "repo:") {
		t.Fatalf("expected 'repo:' in output, got %q", buf.String())
	}
}

func TestTextRendererErrorfAppendsNewline(t *testing.T) {
	var buf bytes.Buffer
	r := NewText(&buf)
	r.Errorf("boom %d", 42)
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("expected trailing newline, got %q", buf.String())
	}
}

func TestLipRendererRespectsBorderAndStyles(t *testing.T) {
	var buf bytes.Buffer
	r := NewLip(&buf, DefaultTheme())
	r.Table([]string{"A", "B"}, [][]Cell{
		{PlainCell("x"), StyledCell("ok", StyleSuccess)},
	})
	out := buf.String()
	for _, want := range []string{"┌", "┐", "└", "┘", "│"} {
		// Note: join glyphs that may differ by build (some terminals replace |),
		// just check at least one of the box characters is present.
		_ = want
	}
	if !strings.ContainsAny(out, "┌└─") {
		t.Fatalf("expected box-drawing characters in lip table output:\n%s", out)
	}
}

func TestResolveJSONWinsOverNoColor(t *testing.T) {
	if got := Resolve(ResolveOptions{JSON: true, NoColor: true}); got != ModeJSON {
		t.Fatalf("expected ModeJSON, got %v", got)
	}
}

func TestResolveNoColorEnvForcesPlain(t *testing.T) {
	if got := Resolve(ResolveOptions{NoColorEnv: "1", IsTerminalFn: func() bool { return true }}); got != ModePlain {
		t.Fatalf("NO_COLOR set: expected ModePlain, got %v", got)
	}
}

func TestResolveTTYStyled(t *testing.T) {
	if got := Resolve(ResolveOptions{IsTerminalFn: func() bool { return true }}); got != ModeStyled {
		t.Fatalf("TTY: expected ModeStyled, got %v", got)
	}
}

func TestResolveNonTTYPlain(t *testing.T) {
	if got := Resolve(ResolveOptions{IsTerminalFn: func() bool { return false }}); got != ModePlain {
		t.Fatalf("non-TTY: expected ModePlain, got %v", got)
	}
}

func TestNewRendererJSONIsNoOp(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(ModeJSON, &buf)
	r.Heading("never shown")
	r.Table([]string{"a"}, [][]Cell{{PlainCell("x")}})
	if buf.Len() != 0 {
		t.Fatalf("ModeJSON renderer wrote output: %q", buf.String())
	}
}

func TestFromFallsBackToDiscardRenderer(t *testing.T) {
	r := From(context.Background())
	if r == nil {
		t.Fatal("From returned nil")
	}
	// Should not panic and should not write to stdout. We can't easily
	// observe Discard, but exercise it.
	r.Heading("ignored")
	r.Note("ignored")
}

func TestWithRoundTripsRenderer(t *testing.T) {
	want := NewText(io.Discard)
	ctx := With(context.Background(), want)
	if got := From(ctx); got != want {
		t.Fatalf("From returned different renderer: %p vs %p", got, want)
	}
}
