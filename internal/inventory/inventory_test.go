package inventory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tzoght/tztoolbox/internal/installer"
	"github.com/tzoght/tztoolbox/internal/model"
)

// scenario builds a fixture root + home covering every status value at once,
// for the cursor tool. Other tools share the same scanner code paths.
//
// Layout produced:
//
//	root/.cursor/commands/in-sync.md          == home
//	root/.cursor/commands/changed.md          != home
//	root/.cursor/commands/only-repo.md        no home counterpart
//	home/.cursor/commands/only-home.md        no repo counterpart
//	root/.cursor/skills/in-sync/SKILL.md      == home tree
//	root/.cursor/skills/changed/SKILL.md      home version is hand-edited
//	home/.cursor/.tztoolbox.manifest          tracks all repo-side paths
//	                                          plus changed.md (so it counts
//	                                          as managed even after edit)
func scenario(t *testing.T) (root, home string) {
	t.Helper()
	root = t.TempDir()
	home = t.TempDir()

	cursorRepo := filepath.Join(root, ".cursor")
	cursorHome := filepath.Join(home, ".cursor")

	mustWrite(t, filepath.Join(cursorRepo, "commands", "in-sync.md"), "same\n")
	mustWrite(t, filepath.Join(cursorHome, "commands", "in-sync.md"), "same\n")

	mustWrite(t, filepath.Join(cursorRepo, "commands", "changed.md"), "v2\n")
	mustWrite(t, filepath.Join(cursorHome, "commands", "changed.md"), "v1\n")

	mustWrite(t, filepath.Join(cursorRepo, "commands", "only-repo.md"), "x\n")
	mustWrite(t, filepath.Join(cursorHome, "commands", "only-home.md"), "y\n")

	// in-sync skill
	mustWrite(t, filepath.Join(cursorRepo, "skills", "in-sync", "SKILL.md"), "skill\n")
	mustWrite(t, filepath.Join(cursorHome, "skills", "in-sync", "SKILL.md"), "skill\n")

	// changed skill: hand-edited inside a sub-file
	mustWrite(t, filepath.Join(cursorRepo, "skills", "changed", "SKILL.md"), "v1\n")
	mustWrite(t, filepath.Join(cursorRepo, "skills", "changed", "references", "ref.md"), "ref\n")
	mustWrite(t, filepath.Join(cursorHome, "skills", "changed", "SKILL.md"), "v1\n")
	mustWrite(t, filepath.Join(cursorHome, "skills", "changed", "references", "ref.md"), "EDITED\n")

	// Manifest declares every path the installer would have written. The
	// `only-home.md` is intentionally omitted so the --managed filter can
	// distinguish it from real installs.
	manifest := []string{
		"commands/in-sync.md",
		"commands/changed.md",
		"commands/only-repo.md",
		"skills/in-sync/SKILL.md",
		"skills/changed/SKILL.md",
		"skills/changed/references/ref.md",
	}
	manifestPath := filepath.Join(cursorHome, installer.ManifestName)
	mustWrite(t, manifestPath, manifestText(manifest))
	return root, home
}

func TestScanProducesAllFourStatusValues(t *testing.T) {
	root, home := scenario(t)
	rep, err := Scan(root, home, []model.Tool{model.Cursor}, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := byNameAndKind(rep.Tools[model.Cursor].Items)

	cases := []struct {
		key  string
		want Status
	}{
		{"command/in-sync", StatusInSync},
		{"command/changed", StatusChanged},
		{"command/only-repo", StatusOnlyInRepo},
		{"command/only-home", StatusOnlyInHome},
		{"skill/in-sync", StatusInSync},
		{"skill/changed", StatusChanged},
	}
	for _, c := range cases {
		it, ok := got[c.key]
		if !ok {
			t.Errorf("missing entry %s", c.key)
			continue
		}
		if it.Status != c.want {
			t.Errorf("%s: got status=%s want=%s", c.key, it.Status, c.want)
		}
	}
}

func TestManagedOnlyHidesUnmanagedHomeFiles(t *testing.T) {
	root, home := scenario(t)
	rep, err := Scan(root, home, []model.Tool{model.Cursor}, Options{ManagedOnly: true})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := byNameAndKind(rep.Tools[model.Cursor].Items)
	if _, exists := got["command/only-home"]; exists {
		t.Errorf("only-home.md was unmanaged but appeared with --managed")
	}
	// Managed items must still show up.
	if _, ok := got["command/in-sync"]; !ok {
		t.Errorf("in-sync.md missing from --managed output")
	}
	for _, it := range rep.Tools[model.Cursor].Items {
		if !it.Managed {
			t.Errorf("--managed leaked unmanaged item: %+v", it)
		}
	}
}

func TestScanReturnsEmptyEntryForMissingHome(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	// Create a repo-side artifact so we know the scan reaches the kind
	// scanners even though the home dir doesn't exist.
	mustWrite(t, filepath.Join(root, ".claude", "commands", "foo.md"), "body\n")

	rep, err := Scan(root, home, []model.Tool{model.Claude}, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	entry := rep.Tools[model.Claude]
	if entry.HomeExists {
		t.Errorf("HomeExists=true for missing dir")
	}
	if len(entry.Items) != 1 {
		t.Fatalf("expected 1 item, got %d (%+v)", len(entry.Items), entry.Items)
	}
	if entry.Items[0].Status != StatusOnlyInRepo {
		t.Errorf("expected only_in_repo, got %s", entry.Items[0].Status)
	}
}

func TestScanRecognisesConcatenatedRuleBundle(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	// Both have a CLAUDE.md but with different bytes.
	mustWrite(t, filepath.Join(root, ".claude", "CLAUDE.md"), "v2\n")
	mustWrite(t, filepath.Join(home, ".claude", "CLAUDE.md"), "v1\n")

	rep, err := Scan(root, home, []model.Tool{model.Claude}, Options{Kinds: []Kind{KindRule}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	items := rep.Tools[model.Claude].Items
	if len(items) != 1 {
		t.Fatalf("expected single rule-bundle entry, got %d (%+v)", len(items), items)
	}
	got := items[0]
	if got.Name != "CLAUDE.md" || got.Kind != KindRule || got.Status != StatusChanged {
		t.Errorf("rule bundle entry wrong: %+v", got)
	}
}

func TestScanCodexUsesPromptsAndAgentsBundle(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	mustWrite(t, filepath.Join(root, ".codex", "prompts", "hi.md"), "hi\n")
	mustWrite(t, filepath.Join(home, ".codex", "prompts", "hi.md"), "hi\n")
	mustWrite(t, filepath.Join(root, ".codex", "AGENTS.md"), "agents\n")
	mustWrite(t, filepath.Join(home, ".codex", "AGENTS.md"), "agents\n")

	rep, err := Scan(root, home, []model.Tool{model.Codex}, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := byNameAndKind(rep.Tools[model.Codex].Items)
	if it, ok := got["command/hi"]; !ok || it.Status != StatusInSync || it.RelPath != "prompts/hi.md" {
		t.Errorf("codex command entry wrong: ok=%t %+v", ok, it)
	}
	if it, ok := got["rule/AGENTS.md"]; !ok || it.Status != StatusInSync {
		t.Errorf("codex agents bundle wrong: ok=%t %+v", ok, it)
	}
}

func TestKindFilterRestrictsResults(t *testing.T) {
	root, home := scenario(t)
	rep, err := Scan(root, home, []model.Tool{model.Cursor}, Options{Kinds: []Kind{KindCommand}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	for _, it := range rep.Tools[model.Cursor].Items {
		if it.Kind != KindCommand {
			t.Errorf("kind filter leaked: %+v", it)
		}
	}
}

// ---------- helpers -------------------------------------------------------

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func manifestText(lines []string) string {
	out := "# tztoolbox-managed paths\n"
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

func byNameAndKind(items []Item) map[string]Item {
	out := map[string]Item{}
	for _, it := range items {
		out[string(it.Kind)+"/"+it.Name] = it
	}
	return out
}
