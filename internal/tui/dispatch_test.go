package tui

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// makeFactory returns a RootCmdFactory that builds a small Cobra tree which
// counts dispatch invocations and lets us assert the captured output.
func makeFactory(buf *strings.Builder) RootCmdFactory {
	return func() *cobra.Command {
		root := &cobra.Command{Use: "tzcli"}
		root.AddCommand(&cobra.Command{
			Use: "echo",
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.Println(strings.Join(args, " "))
				return nil
			},
		})
		root.SetOut(buf)
		root.SetErr(buf)
		return root
	}
}

func TestRunLineHappyPath(t *testing.T) {
	var buf strings.Builder
	m := NewModel(Options{Factory: makeFactory(&buf)})
	cmd := m.runLine("echo hello world")
	msg := cmd().(commandDoneMsg)
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if !strings.Contains(msg.output, "hello world") {
		t.Fatalf("expected output to contain 'hello world', got %q", msg.output)
	}
	if msg.head != "echo" {
		t.Fatalf("expected head=echo, got %q", msg.head)
	}
	if msg.duration <= 0 {
		t.Fatalf("expected non-zero duration, got %v", msg.duration)
	}
}

func TestRunLineExpandsMnemonic(t *testing.T) {
	// Build a factory whose root has a `sync` subcommand so the mnemonic
	// "s" should be routed to it after expansion.
	factory := func() *cobra.Command {
		root := &cobra.Command{Use: "tzcli"}
		root.AddCommand(&cobra.Command{
			Use: "sync",
			RunE: func(c *cobra.Command, _ []string) error {
				c.Println("synced")
				return nil
			},
		})
		var buf strings.Builder
		root.SetOut(&buf)
		root.SetErr(&buf)
		return root
	}
	m := NewModel(Options{Factory: factory})
	msg := m.runLine("s")().(commandDoneMsg)
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if !strings.Contains(msg.output, "synced") {
		t.Fatalf("expected mnemonic 's' to invoke sync, got %q", msg.output)
	}
	if msg.head != "sync" {
		t.Fatalf("expected head=sync after expansion, got %q", msg.head)
	}
}

func TestRunLineUnknownCommandReportsError(t *testing.T) {
	var buf strings.Builder
	m := NewModel(Options{Factory: makeFactory(&buf)})
	cmd := m.runLine("nope --foo")
	msg := cmd().(commandDoneMsg)
	if msg.err == nil {
		t.Fatalf("expected non-nil error, got output %q", msg.output)
	}
}

func TestRunLineParseError(t *testing.T) {
	var buf strings.Builder
	m := NewModel(Options{Factory: makeFactory(&buf)})
	// Unterminated quote should fail at the shell parser.
	cmd := m.runLine(`echo "unterminated`)
	msg := cmd().(commandDoneMsg)
	if msg.err == nil {
		t.Fatalf("expected parse error")
	}
	if !strings.Contains(msg.err.Error(), "parse error") {
		t.Fatalf("expected 'parse error' in message, got %q", msg.err.Error())
	}
}

func TestRunLineStripsLeadingTzcli(t *testing.T) {
	var buf strings.Builder
	m := NewModel(Options{Factory: makeFactory(&buf)})
	cmd := m.runLine("tzcli echo ok")
	msg := cmd().(commandDoneMsg)
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if !strings.Contains(msg.output, "ok") {
		t.Fatalf("expected output to contain 'ok', got %q", msg.output)
	}
}

func TestFlagJSONFromArgsDetectsVariants(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	cases := map[string]bool{
		"plain":          false,
		"--json":         true,
		"--json=true":    true,
		"--no-color":     false,
		"installed":      false,
		"installed-json": false,
	}
	for arg, want := range cases {
		got := m.flagJSONFromArgs([]string{arg})
		if got != want {
			t.Fatalf("flagJSONFromArgs(%q): want %v, got %v", arg, want, got)
		}
	}
}
