package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// fakeRoot returns a small cobra tree resembling the real tzcli surface,
// just enough to exercise completion logic.
func fakeRoot() *cobra.Command {
	root := &cobra.Command{Use: "tzcli"}
	sync := &cobra.Command{Use: "sync"}
	sync.Flags().Bool("check", false, "")
	sync.Flags().Bool("dry-run", false, "")
	sync.Flags().String("only", "all", "")

	syrup := &cobra.Command{Use: "syrup"} // shares prefix with sync
	doctor := &cobra.Command{Use: "doctor"}
	doctor.Flags().Bool("verbose", false, "")

	hidden := &cobra.Command{Use: "hidden", Hidden: true}

	root.AddCommand(sync, syrup, doctor, hidden)
	return root
}

func TestCompleteEmptyListsAllSubcommands(t *testing.T) {
	root := fakeRoot()
	_, suggestions := completeLine(root, "")
	want := []string{"doctor", "sync", "syrup"}
	if !reflect.DeepEqual(suggestions, want) {
		t.Fatalf("want %v, got %v", want, suggestions)
	}
}

func TestCompleteUniquePrefixSubcommand(t *testing.T) {
	root := fakeRoot()
	got, _ := completeLine(root, "do")
	if got != "doctor " {
		t.Fatalf("expected 'doctor ', got %q", got)
	}
}

func TestCompleteAmbiguousPrefixExtendsToCommonPrefix(t *testing.T) {
	root := fakeRoot()
	got, suggestions := completeLine(root, "sy")
	if got != "sy" {
		t.Fatalf("expected unchanged 'sy' (already at common prefix), got %q", got)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %v", suggestions)
	}
}

func TestCompleteSubcommandFlag(t *testing.T) {
	root := fakeRoot()
	got, _ := completeLine(root, "sync --che")
	if got != "sync --check " {
		t.Fatalf("expected 'sync --check ', got %q", got)
	}
}

func TestCompleteIgnoresHiddenCommand(t *testing.T) {
	root := fakeRoot()
	_, suggestions := completeLine(root, "hi")
	if len(suggestions) != 0 {
		t.Fatalf("hidden command should not suggest, got %v", suggestions)
	}
}

func TestStripBareNameDropsLeadingTzcli(t *testing.T) {
	if got := stripBareName([]string{"tzcli", "sync"}); !reflect.DeepEqual(got, []string{"sync"}) {
		t.Fatalf("stripBareName: %v", got)
	}
	if got := stripBareName([]string{"sync"}); !reflect.DeepEqual(got, []string{"sync"}) {
		t.Fatalf("stripBareName: %v", got)
	}
}

func TestCompleteFlagPrefixUnchangedAtCommonPrefix(t *testing.T) {
	root := fakeRoot()
	got, suggestions := completeLine(root, "sync --")
	if !strings.HasPrefix(got, "sync --") {
		t.Fatalf("expected line to keep '--' prefix, got %q", got)
	}
	if len(suggestions) < 2 {
		t.Fatalf("expected multiple flag suggestions, got %v", suggestions)
	}
}
