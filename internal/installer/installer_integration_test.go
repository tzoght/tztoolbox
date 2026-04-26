//go:build integration

package installer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tzoght/tztoolbox/internal/installer"
	"github.com/tzoght/tztoolbox/internal/inventory"
	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/render"
)

// TestInstallEndToEnd builds a fixture repo, renders it, installs into a
// temp HOME, and verifies all three tool dirs end up populated and the
// manifest accurately reflects the installed paths.
func TestInstallEndToEnd(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	// Pre-create all three tool home dirs so install auto-detection picks
	// up every tool (matching what a real workstation looks like).
	for _, tool := range model.AllTools {
		if err := os.MkdirAll(filepath.Join(home, "."+string(tool)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite(t, filepath.Join(root, "go.mod"), "module github.com/tzoght/tztoolbox\n")
	mustWrite(t, filepath.Join(root, "shared", "commands", "hi.md"), "# hi\nhello\n")
	mustWrite(t, filepath.Join(root, "shared", "rules", "r1.md"), "# rule\nbody\n")
	mustWrite(t, filepath.Join(root, "shared", "skills", "demo", "SKILL.md"),
		"---\nname: demo\ndescription: demo skill\n---\n# demo\nbody\n")
	mustWrite(t, filepath.Join(root, "overrides", "cursor", "rules", "r1.frontmatter.yaml"),
		"alwaysApply: true\n")

	if _, err := render.Render(root, render.Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}

	res, err := installer.Install(root, installer.Options{HomeDir: home, Prune: true})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	for _, tool := range model.AllTools {
		tr := res.PerTool[tool]
		if !tr.Installed {
			t.Errorf("%s not installed: %s", tool, tr.Skipped)
			continue
		}
		// Manifest must exist and list every copied path.
		manifest := filepath.Join(tr.HomePath, installer.ManifestName)
		if _, err := os.Stat(manifest); err != nil {
			t.Errorf("%s manifest missing: %v", tool, err)
		}
	}

	// Cursor command landed under commands/.
	if _, err := os.Stat(filepath.Join(home, ".cursor", "commands", "hi.md")); err != nil {
		t.Errorf("cursor command missing: %v", err)
	}
	// Codex command landed under prompts/.
	if _, err := os.Stat(filepath.Join(home, ".codex", "prompts", "hi.md")); err != nil {
		t.Errorf("codex prompt missing: %v", err)
	}
	// Claude rules landed in CLAUDE.md.
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md missing: %v", err)
	}

	// Contract with the inventory package: every managed entry should be
	// in_sync immediately after a fresh install. This catches manifest
	// drift (e.g. installer writes a path inventory doesn't recognise) and
	// status-calculation drift in either direction.
	rep, err := inventory.Scan(root, home, model.AllTools, inventory.Options{ManagedOnly: true})
	if err != nil {
		t.Fatalf("inventory.Scan: %v", err)
	}
	totalManaged := 0
	for tool, entry := range rep.Tools {
		if !entry.HomeExists {
			t.Errorf("%s home should exist after install", tool)
		}
		for _, it := range entry.Items {
			totalManaged++
			if it.Status != inventory.StatusInSync {
				t.Errorf("post-install drift on %s/%s/%s: status=%s", tool, it.Kind, it.Name, it.Status)
			}
			if !it.Managed {
				t.Errorf("--managed entry not marked managed: %+v", it)
			}
		}
	}
	if totalManaged == 0 {
		t.Fatal("expected at least one managed inventory entry after install")
	}
}

// TestInstallPrunesRemovedArtifacts asserts that re-running install after
// deleting an artifact upstream removes its installed copy.
func TestInstallPrunesRemovedArtifacts(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	for _, tool := range model.AllTools {
		if err := os.MkdirAll(filepath.Join(home, "."+string(tool)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite(t, filepath.Join(root, "shared", "commands", "a.md"), "# a\n")
	mustWrite(t, filepath.Join(root, "shared", "commands", "b.md"), "# b\n")
	mustWrite(t, filepath.Join(root, "overrides", "cursor", "rules", ".keep"), "")

	if _, err := render.Render(root, render.Options{Tools: []model.Tool{model.Cursor}}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if _, err := installer.Install(root, installer.Options{
		HomeDir: home,
		Tools:   []model.Tool{model.Cursor},
		Prune:   true,
	}); err != nil {
		t.Fatalf("Install (1): %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "commands", "a.md")); err != nil {
		t.Fatalf("a.md should be installed: %v", err)
	}

	// Now drop a.md upstream, re-render, re-install.
	if err := os.Remove(filepath.Join(root, "shared", "commands", "a.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := render.Render(root, render.Options{Tools: []model.Tool{model.Cursor}}); err != nil {
		t.Fatalf("Render (2): %v", err)
	}
	if _, err := installer.Install(root, installer.Options{
		HomeDir: home,
		Tools:   []model.Tool{model.Cursor},
		Prune:   true,
	}); err != nil {
		t.Fatalf("Install (2): %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "commands", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("a.md should be pruned, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "commands", "b.md")); err != nil {
		t.Fatalf("b.md should still be installed: %v", err)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
