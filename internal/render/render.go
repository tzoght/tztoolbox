// Package render translates the canonical sources under shared/ (plus any
// per-tool overrides under overrides/<tool>/) into each tool's native tree
// (e.g. .cursor/, .claude/, .codex/) inside the repo.
//
// The output is deterministic: running render twice on unchanged inputs
// produces byte-identical outputs (idempotent). This is what `tzcli sync`
// uses; `tzcli install` simply copies the rendered native trees into the
// user's tool home directories.
package render

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/repo"
)

// Options controls a render pass.
type Options struct {
	// Tools restricts rendering to a subset; empty = all.
	Tools []model.Tool
	// DryRun reports the work without writing anything.
	DryRun bool
}

// Result summarizes what a render pass did.
type Result struct {
	// PerTool counts files written per tool, keyed by tool name.
	PerTool map[model.Tool]ToolCounts
}

// ToolCounts is a tally of artifacts emitted for one tool.
type ToolCounts struct {
	Commands int
	Rules    int
	Skills   int
}

// Disabled lists artifact names a tool does NOT want included.
type Disabled struct {
	Commands []string `yaml:"commands"`
	Rules    []string `yaml:"rules"`
	Skills   []string `yaml:"skills"`
}

// Render reads <root>/shared and <root>/overrides and writes the per-tool
// native trees into <root>/.<tool>. It is safe to call repeatedly.
func Render(root string, opts Options) (Result, error) {
	tools := opts.Tools
	if len(tools) == 0 {
		tools = model.AllTools
	}
	res := Result{PerTool: map[model.Tool]ToolCounts{}}

	for _, t := range tools {
		c, err := renderTool(root, t, opts.DryRun)
		if err != nil {
			return res, fmt.Errorf("render %s: %w", t, err)
		}
		res.PerTool[t] = c
	}
	return res, nil
}

func renderTool(root string, t model.Tool, dryRun bool) (ToolCounts, error) {
	var counts ToolCounts

	disabled, err := loadDisabled(root, t)
	if err != nil {
		return counts, err
	}

	dest := filepath.Join(root, t.RepoSubdir())
	if !dryRun {
		if err := resetToolDir(dest); err != nil {
			return counts, fmt.Errorf("reset %s: %w", dest, err)
		}
	}

	cmds, err := renderCommands(root, t, dest, disabled.Commands, dryRun)
	if err != nil {
		return counts, err
	}
	counts.Commands = cmds

	rules, err := renderRules(root, t, dest, disabled.Rules, dryRun)
	if err != nil {
		return counts, err
	}
	counts.Rules = rules

	skills, err := renderSkills(root, t, dest, disabled.Skills, dryRun)
	if err != nil {
		return counts, err
	}
	counts.Skills = skills

	return counts, nil
}

// resetToolDir removes any prior content for this tool's in-repo native dir,
// then recreates it. This guarantees `tzcli sync` is fully deterministic and
// no stale files survive a deletion in shared/.
func resetToolDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// loadDisabled reads overrides/<tool>/disabled.yaml if present.
func loadDisabled(root string, t model.Tool) (Disabled, error) {
	var d Disabled
	p := filepath.Join(repo.OverridesDir(root), t.String(), "disabled.yaml")
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, fmt.Errorf("read %s: %w", p, err)
	}
	if err := yaml.Unmarshal(b, &d); err != nil {
		return d, fmt.Errorf("parse %s: %w", p, err)
	}
	return d, nil
}

// ----- commands ----------------------------------------------------------

func renderCommands(root string, t model.Tool, toolDir string, disabled []string, dryRun bool) (int, error) {
	srcDir := repo.SharedCommandsDir(root)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read %s: %w", srcDir, err)
	}
	dstDir := filepath.Join(toolDir, t.CommandsSubdir())
	if !dryRun {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return 0, err
		}
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".md")
		if isDisabled(base, disabled) {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(dstDir, base+t.CommandFileExt())
		if !dryRun {
			if err := copyFile(src, dst); err != nil {
				return count, fmt.Errorf("copy command %s: %w", e.Name(), err)
			}
		}
		count++
	}
	return count, nil
}

// ----- rules -------------------------------------------------------------

func renderRules(root string, t model.Tool, toolDir string, disabled []string, dryRun bool) (int, error) {
	srcDir := repo.SharedRulesDir(root)
	names, err := listMarkdownNames(srcDir)
	if err != nil {
		return 0, err
	}

	switch t.RulesMode() {
	case model.RulesAsFiles:
		return renderRulesAsFiles(root, t, toolDir, names, disabled, dryRun)
	case model.RulesConcatenated:
		return renderRulesConcatenated(srcDir, t, toolDir, names, disabled, dryRun)
	}
	return 0, nil
}

func renderRulesAsFiles(root string, t model.Tool, toolDir string, names, disabled []string, dryRun bool) (int, error) {
	dstDir := filepath.Join(toolDir, t.RulesPath())
	if !dryRun {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return 0, err
		}
	}
	count := 0
	for _, base := range names {
		if isDisabled(base, disabled) {
			continue
		}
		bodyPath := filepath.Join(repo.SharedRulesDir(root), base+".md")
		body, err := os.ReadFile(bodyPath)
		if err != nil {
			return count, err
		}
		fmYAML, err := readCursorRuleFrontmatter(root, base)
		if err != nil {
			return count, err
		}
		dst := filepath.Join(dstDir, base+t.RuleFileExt())
		out := composeMDC(fmYAML, body)
		if !dryRun {
			if err := writeFile(dst, out); err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}

func renderRulesConcatenated(srcDir string, t model.Tool, toolDir string, names, disabled []string, dryRun bool) (int, error) {
	var sb strings.Builder
	// Generation banner, kept as an HTML comment so it doesn't disturb the
	// markdown heading hierarchy that follows.
	sb.WriteString("<!-- generated by `tzcli sync`; edit shared/rules/*.md instead -->\n\n")

	count := 0
	for _, base := range names {
		if isDisabled(base, disabled) {
			continue
		}
		body, err := os.ReadFile(filepath.Join(srcDir, base+".md"))
		if err != nil {
			return count, err
		}
		if count > 0 {
			sb.WriteString("\n---\n\n")
		}
		// Preserve the body verbatim: its own H1 acts as the section title,
		// and downstream `## ...` headings keep their natural depth.
		bs := strings.TrimRight(string(body), "\n")
		sb.WriteString(bs)
		sb.WriteString("\n")
		count++
	}

	if !dryRun {
		dst := filepath.Join(toolDir, t.RulesPath())
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return count, err
		}
		if err := writeFile(dst, []byte(sb.String())); err != nil {
			return count, err
		}
	}
	return count, nil
}

func readCursorRuleFrontmatter(root, base string) ([]byte, error) {
	p := filepath.Join(repo.CursorOverridesDir(root), "rules", base+".frontmatter.yaml")
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return b, err
}

// composeMDC writes a Cursor `.mdc` file: optional YAML frontmatter delimited
// by `---` markers, then the markdown body. The output always ends with a
// single trailing newline.
func composeMDC(frontmatterYAML, body []byte) []byte {
	var out strings.Builder
	if len(frontmatterYAML) > 0 {
		out.WriteString("---\n")
		out.Write(frontmatterYAML)
		if len(frontmatterYAML) > 0 && frontmatterYAML[len(frontmatterYAML)-1] != '\n' {
			out.WriteByte('\n')
		}
		out.WriteString("---\n\n")
	}
	out.Write(body)
	if len(body) == 0 || body[len(body)-1] != '\n' {
		out.WriteByte('\n')
	}
	return []byte(out.String())
}

// ----- skills ------------------------------------------------------------

func renderSkills(root string, t model.Tool, toolDir string, disabled []string, dryRun bool) (int, error) {
	srcDir := repo.SharedSkillsDir(root)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	dstDir := filepath.Join(toolDir, t.SkillsSubdir())
	if !dryRun {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return 0, err
		}
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if isDisabled(e.Name(), disabled) {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(dstDir, e.Name())
		if !dryRun {
			if err := copyDir(src, dst); err != nil {
				return count, fmt.Errorf("copy skill %s: %w", e.Name(), err)
			}
		}
		count++
	}
	return count, nil
}

// ----- helpers -----------------------------------------------------------

func listMarkdownNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		out = append(out, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(out)
	return out, nil
}

func isDisabled(name string, disabled []string) bool {
	for _, d := range disabled {
		if d == name {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		return copyFile(p, target)
	})
}
