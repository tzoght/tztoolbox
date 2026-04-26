package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tzoght/tztoolbox/internal/model"
)

// fixtureRoot builds a minimal tztoolbox repo under t.TempDir() so the
// renderer can run end-to-end without depending on the real workspace.
func fixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// shared/commands/hello.md
	mustWrite(t, filepath.Join(root, "shared", "commands", "hello.md"), "# hello\n\nbody\n")

	// shared/rules/sample.md (with H1)
	mustWrite(t, filepath.Join(root, "shared", "rules", "sample.md"), "# Sample rule\n\n## Sub\n\nbody\n")

	// shared/skills/greeter/SKILL.md
	skill := `---
name: greeter
description: This skill should be used when the user wants a greeting.
---

# greeter

body
`
	mustWrite(t, filepath.Join(root, "shared", "skills", "greeter", "SKILL.md"), skill)

	// overrides/cursor/rules/sample.frontmatter.yaml
	mustWrite(t, filepath.Join(root, "overrides", "cursor", "rules", "sample.frontmatter.yaml"), "alwaysApply: true\n")

	// go.mod marker so repo.Discover would work if needed
	mustWrite(t, filepath.Join(root, "go.mod"), "module github.com/tzoght/tztoolbox\n\ngo 1.22\n")

	return root
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

func TestRenderEmitsAllToolsWithCorrectLayout(t *testing.T) {
	root := fixtureRoot(t)

	res, err := Render(root, Options{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// Each tool sees the same number of artifacts.
	for _, tool := range model.AllTools {
		got := res.PerTool[tool]
		if got.Commands != 1 || got.Skills != 1 {
			t.Errorf("%s counts: %+v; want commands=1 skills=1", tool, got)
		}
	}

	// Cursor: rule is its own .mdc file with frontmatter prepended.
	mdc := readFile(t, filepath.Join(root, ".cursor", "rules", "sample.mdc"))
	if !strings.HasPrefix(mdc, "---\nalwaysApply: true\n---\n\n# Sample rule") {
		t.Errorf(".cursor/rules/sample.mdc has unexpected prefix:\n%s", mdc)
	}

	// Claude: rules concatenated into CLAUDE.md.
	cl := readFile(t, filepath.Join(root, ".claude", "CLAUDE.md"))
	if !strings.Contains(cl, "# Sample rule") {
		t.Errorf("CLAUDE.md missing rule body:\n%s", cl)
	}
	if strings.Contains(cl, "alwaysApply") {
		t.Errorf("CLAUDE.md should not include Cursor frontmatter")
	}

	// Codex: rules concatenated into AGENTS.md; commands under prompts/.
	if !fileExists(filepath.Join(root, ".codex", "AGENTS.md")) {
		t.Error(".codex/AGENTS.md missing")
	}
	if !fileExists(filepath.Join(root, ".codex", "prompts", "hello.md")) {
		t.Error(".codex/prompts/hello.md missing (Codex uses prompts/ not commands/)")
	}

	// Skill copied verbatim into all three.
	for _, tool := range model.AllTools {
		p := filepath.Join(root, "."+string(tool), "skills", "greeter", "SKILL.md")
		if !fileExists(p) {
			t.Errorf("skill missing for %s: %s", tool, p)
		}
	}
}

func TestRenderIsIdempotent(t *testing.T) {
	root := fixtureRoot(t)

	if _, err := Render(root, Options{}); err != nil {
		t.Fatalf("first render: %v", err)
	}
	first := snapshot(t, root)

	if _, err := Render(root, Options{}); err != nil {
		t.Fatalf("second render: %v", err)
	}
	second := snapshot(t, root)

	if len(first) != len(second) {
		t.Fatalf("file count differs: first=%d second=%d", len(first), len(second))
	}
	for path, b := range first {
		if string(second[path]) != string(b) {
			t.Errorf("%s changed between identical renders", path)
		}
	}
}

func TestCheckDriftReportsNothingAfterFreshRender(t *testing.T) {
	root := fixtureRoot(t)
	if _, err := Render(root, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	rep, err := CheckDrift(root, model.AllTools)
	if err != nil {
		t.Fatalf("CheckDrift: %v", err)
	}
	if rep.HasDrift() {
		t.Fatalf("expected no drift, got %+v", rep)
	}
}

func TestCheckDriftReportsHandEdit(t *testing.T) {
	root := fixtureRoot(t)
	if _, err := Render(root, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Hand-edit a generated file.
	mustWrite(t, filepath.Join(root, ".cursor", "commands", "hello.md"), "TAMPERED\n")

	rep, err := CheckDrift(root, model.AllTools)
	if err != nil {
		t.Fatalf("CheckDrift: %v", err)
	}
	if !rep.HasDrift() {
		t.Fatal("expected drift after hand edit, got none")
	}
}

func TestRenderHonorsDisabled(t *testing.T) {
	root := fixtureRoot(t)
	mustWrite(t, filepath.Join(root, "overrides", "claude", "disabled.yaml"),
		"commands:\n  - hello\n")

	if _, err := Render(root, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if fileExists(filepath.Join(root, ".claude", "commands", "hello.md")) {
		t.Error("disabled command was still rendered for Claude")
	}
	if !fileExists(filepath.Join(root, ".cursor", "commands", "hello.md")) {
		t.Error("Cursor render dropped command that was only disabled for Claude")
	}
}

func snapshot(t *testing.T, root string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, tool := range model.AllTools {
		dir := filepath.Join(root, "."+string(tool))
		_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			b, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(root, p)
			out[rel] = b
			return nil
		})
	}
	return out
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
