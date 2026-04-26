package tui

import (
	"strings"
	"testing"
)

func TestExpandMnemonicSingleLetter(t *testing.T) {
	cases := map[string]string{
		"s":            "sync",
		"d":            "doctor",
		"i":            "install",
		"I":            "installed",
		"x":            "disable",
		"/":            "search",
		"?":            "help",
		"q":            "quit",
		"a --help":     "add --help",
		"s --check":    "sync --check",
		"  s  --check": "sync  --check",
	}
	for in, want := range cases {
		if got := expandMnemonic(in); got != want {
			t.Errorf("expandMnemonic(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpandMnemonicLeavesFullCommandsAlone(t *testing.T) {
	cases := []string{"sync", "sync --check", "doctor", "help add", ""}
	for _, in := range cases {
		if got := expandMnemonic(in); got != in {
			t.Errorf("expandMnemonic(%q) modified input: %q", in, got)
		}
	}
}

func TestExpandMnemonicUnknownLetter(t *testing.T) {
	in := "z --foo"
	if got := expandMnemonic(in); got != in {
		t.Errorf("expandMnemonic(%q) = %q, expected unchanged", in, got)
	}
}

func TestRenderStartupMenuContainsAllMnemonics(t *testing.T) {
	out := renderStartupMenu(defaultChromeStyles(), "v0.0.0-test")
	if !strings.Contains(out, "welcome to tzcli v0.0.0-test") {
		t.Fatalf("missing welcome line in:\n%s", out)
	}
	for _, mn := range mnemonics {
		if !strings.Contains(out, mn.Cmd) {
			t.Errorf("startup menu missing command %q", mn.Cmd)
		}
	}
}

func TestMnemonicKeysAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, mn := range mnemonics {
		if existing, ok := seen[mn.Key]; ok {
			t.Errorf("mnemonic key %q reused for %q (already used by %q)", mn.Key, mn.Cmd, existing)
		}
		seen[mn.Key] = mn.Cmd
	}
}

func TestMnemonicCommandsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, mn := range mnemonics {
		if seen[mn.Cmd] {
			t.Errorf("mnemonic command %q listed twice", mn.Cmd)
		}
		seen[mn.Cmd] = true
	}
}

func TestHelpHintForKnownCommand(t *testing.T) {
	got := helpHint("add")
	if !strings.Contains(got, "add --help") || !strings.Contains(got, "help add") {
		t.Fatalf("unexpected hint: %q", got)
	}
}

func TestHelpHintSuppressedForBuiltins(t *testing.T) {
	for _, head := range []string{"", "help", "quit", "exit", "clear", ":q"} {
		if got := helpHint(head); got != "" {
			t.Errorf("expected empty hint for %q, got %q", head, got)
		}
	}
}

func TestMnemonicPaletteItemsIncludeKeys(t *testing.T) {
	items := mnemonicPaletteItems()
	if len(items) != len(mnemonics) {
		t.Fatalf("expected %d items, got %d", len(mnemonics), len(items))
	}
	for _, it := range items {
		if !strings.Contains(it.desc, "[") {
			t.Errorf("palette item %q missing mnemonic prefix in desc: %q", it.command, it.desc)
		}
	}
}
