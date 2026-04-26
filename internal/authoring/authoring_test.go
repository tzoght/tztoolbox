package authoring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitFrontmatter(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		wantFM      string
		wantBody    string
		expectError bool
	}{
		{
			name:     "plain frontmatter",
			in:       "---\nname: foo\n---\nbody\n",
			wantFM:   "name: foo",
			wantBody: "body\n",
		},
		{
			name:     "no frontmatter",
			in:       "no fm here\n",
			wantFM:   "",
			wantBody: "no fm here\n",
		},
		{
			name:        "unterminated frontmatter",
			in:          "---\nname: foo\nbody\n",
			expectError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fm, body, err := splitFrontmatter([]byte(tc.in))
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := strings.TrimSpace(string(fm)); got != tc.wantFM {
				t.Errorf("frontmatter: got %q want %q", got, tc.wantFM)
			}
			if string(body) != tc.wantBody {
				t.Errorf("body: got %q want %q", string(body), tc.wantBody)
			}
		})
	}
}

func TestValidateNameRejectsBad(t *testing.T) {
	bad := []string{"", "a", "-foo", "foo-", "Foo", "foo bar", strings.Repeat("a", 100)}
	for _, n := range bad {
		if err := validateName(n); err == nil {
			t.Errorf("validateName(%q) should error", n)
		}
	}
	good := []string{"foo", "foo-bar", "ab", "f1", "tz-branch-create"}
	for _, n := range good {
		if err := validateName(n); err != nil {
			t.Errorf("validateName(%q) unexpected error: %v", n, err)
		}
	}
}

func TestAddCommandWritesScaffold(t *testing.T) {
	root := scratchRepo(t)
	created, err := Add(root, KindCommand, "tz-demo")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 file created, got %d: %v", len(created), created)
	}
	p := filepath.Join(root, created[0])
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Tz Demo") {
		t.Errorf("scaffold missing humanized title; got:\n%s", string(b))
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(root, KindCommand, "tz-demo"); err == nil {
		t.Error("expected refuse-overwrite on second Add")
	}
}

func TestAddRuleWritesBodyAndFrontmatter(t *testing.T) {
	root := scratchRepo(t)
	created, err := Add(root, KindRule, "demo-rule")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("rule scaffold should produce body + frontmatter (got %d: %v)", len(created), created)
	}
}

func TestValidateFlagsBadSkill(t *testing.T) {
	root := scratchRepo(t)
	bad := `---
name: greeter
description: Use this skill when ...
---

no h1 here
`
	if err := os.MkdirAll(filepath.Join(root, "shared", "skills", "greeter"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared", "skills", "greeter", "SKILL.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	issues, err := Validate(root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	gotThirdPersonWarn := false
	for _, i := range issues {
		if strings.Contains(i.Message, "third-person") {
			gotThirdPersonWarn = true
		}
	}
	if !gotThirdPersonWarn {
		t.Errorf("expected third-person warning, got: %+v", issues)
	}
}

func scratchRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"shared/commands", "shared/rules", "shared/skills", "overrides/cursor/rules"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/tzoght/tztoolbox\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
