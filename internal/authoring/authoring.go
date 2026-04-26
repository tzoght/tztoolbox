// Package authoring provides scaffolding for new commands, rules, and skills,
// plus listing and validation of the existing corpus under shared/.
package authoring

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/tzoght/tztoolbox/internal/repo"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Kind names the artifact category to scaffold or list.
type Kind string

const (
	KindCommand Kind = "command"
	KindRule    Kind = "rule"
	KindSkill   Kind = "skill"
)

// AllKinds is a deterministic list of artifact kinds.
var AllKinds = []Kind{KindCommand, KindRule, KindSkill}

// ParseKind validates a user-supplied kind string.
func ParseKind(s string) (Kind, error) {
	switch Kind(strings.ToLower(s)) {
	case KindCommand, KindRule, KindSkill:
		return Kind(strings.ToLower(s)), nil
	}
	return "", fmt.Errorf("unknown kind %q (expected command|rule|skill)", s)
}

// Add scaffolds a new artifact of `kind` named `name` under shared/. It
// returns the list of paths created (relative to root) so the caller can
// surface them to the user.
func Add(root string, kind Kind, name string) ([]string, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	data := struct {
		Name  string
		Title string
	}{
		Name:  name,
		Title: humanize(name),
	}

	switch kind {
	case KindCommand:
		out := filepath.Join(repo.SharedCommandsDir(root), name+".md")
		if err := writeFromTemplate(out, "templates/command.md.tmpl", data); err != nil {
			return nil, err
		}
		return []string{rel(root, out)}, nil

	case KindRule:
		body := filepath.Join(repo.SharedRulesDir(root), name+".md")
		if err := writeFromTemplate(body, "templates/rule.md.tmpl", data); err != nil {
			return nil, err
		}
		fm := filepath.Join(repo.CursorOverridesDir(root), "rules", name+".frontmatter.yaml")
		if err := writeFromTemplate(fm, "templates/rule.frontmatter.yaml.tmpl", data); err != nil {
			return nil, err
		}
		return []string{rel(root, body), rel(root, fm)}, nil

	case KindSkill:
		out := filepath.Join(repo.SharedSkillsDir(root), name, "SKILL.md")
		if err := writeFromTemplate(out, "templates/skill.md.tmpl", data); err != nil {
			return nil, err
		}
		return []string{rel(root, out)}, nil
	}
	return nil, fmt.Errorf("unsupported kind %q", kind)
}

// Item describes one artifact under shared/, used by `list` and `validate`.
type Item struct {
	Kind Kind
	Name string
	Path string // absolute
	Size int64
}

// List returns artifacts of one or more kinds. An empty kinds slice = all.
func List(root string, kinds []Kind) ([]Item, error) {
	if len(kinds) == 0 {
		kinds = AllKinds
	}
	var out []Item
	for _, k := range kinds {
		items, err := listOne(root, k)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func listOne(root string, k Kind) ([]Item, error) {
	switch k {
	case KindCommand:
		return listMarkdownDir(root, repo.SharedCommandsDir(root), KindCommand)
	case KindRule:
		return listMarkdownDir(root, repo.SharedRulesDir(root), KindRule)
	case KindSkill:
		return listSkills(root, repo.SharedSkillsDir(root))
	}
	return nil, nil
}

func listMarkdownDir(root, dir string, kind Kind) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, Item{
			Kind: kind,
			Name: strings.TrimSuffix(e.Name(), ".md"),
			Path: full,
			Size: info.Size(),
		})
	}
	_ = root // unused for these kinds
	return out, nil
}

func listSkills(root, dir string) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := filepath.Join(dir, e.Name())
		var size int64
		_ = filepath.WalkDir(skillDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if info, ierr := d.Info(); ierr == nil {
				size += info.Size()
			}
			return nil
		})
		out = append(out, Item{Kind: KindSkill, Name: e.Name(), Path: skillDir, Size: size})
	}
	_ = root
	return out, nil
}

// Issue is a single problem found by Validate.
type Issue struct {
	Path     string
	Severity string // "error" | "warning"
	Message  string
}

// Validate inspects the corpus and returns a list of issues.
func Validate(root string) ([]Issue, error) {
	var issues []Issue

	skillIssues, err := validateSkills(root)
	if err != nil {
		return issues, err
	}
	issues = append(issues, skillIssues...)

	cmdIssues, err := validateMarkdownH1(repo.SharedCommandsDir(root), "command")
	if err != nil {
		return issues, err
	}
	issues = append(issues, cmdIssues...)

	ruleIssues, err := validateMarkdownH1(repo.SharedRulesDir(root), "rule")
	if err != nil {
		return issues, err
	}
	issues = append(issues, ruleIssues...)

	secretIssues, err := validateNoSecrets(root)
	if err != nil {
		return issues, err
	}
	issues = append(issues, secretIssues...)

	return issues, nil
}

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func validateSkills(root string) ([]Issue, error) {
	dir := repo.SharedSkillsDir(root)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var issues []Issue
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMD := filepath.Join(dir, e.Name(), "SKILL.md")
		b, err := os.ReadFile(skillMD)
		if err != nil {
			issues = append(issues, Issue{Path: skillMD, Severity: "error", Message: "missing SKILL.md"})
			continue
		}
		fm, _, err := splitFrontmatter(b)
		if err != nil {
			issues = append(issues, Issue{Path: skillMD, Severity: "error", Message: err.Error()})
			continue
		}
		var meta skillFrontmatter
		if err := yaml.Unmarshal(fm, &meta); err != nil {
			issues = append(issues, Issue{Path: skillMD, Severity: "error", Message: "frontmatter parse: " + err.Error()})
			continue
		}
		if meta.Name == "" {
			issues = append(issues, Issue{Path: skillMD, Severity: "error", Message: "frontmatter missing `name`"})
		} else if meta.Name != e.Name() {
			issues = append(issues, Issue{Path: skillMD, Severity: "warning", Message: fmt.Sprintf("frontmatter name %q does not match dir %q", meta.Name, e.Name())})
		}
		if meta.Description == "" {
			issues = append(issues, Issue{Path: skillMD, Severity: "error", Message: "frontmatter missing `description`"})
		} else {
			if len(meta.Description) > 200 {
				issues = append(issues, Issue{Path: skillMD, Severity: "warning", Message: fmt.Sprintf("description is %d chars; recommend <=200 for cross-tool compatibility", len(meta.Description))})
			}
			low := strings.ToLower(meta.Description)
			if strings.HasPrefix(low, "use this skill") || strings.HasPrefix(low, "use when") {
				issues = append(issues, Issue{Path: skillMD, Severity: "warning", Message: "description should be third-person (\"This skill should be used when...\") for Claude.ai compatibility"})
			}
		}
	}
	return issues, nil
}

var h1Re = regexp.MustCompile(`(?m)^# .+$`)

func validateMarkdownH1(dir, label string) ([]Issue, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var issues []Issue
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(full)
		if err != nil {
			issues = append(issues, Issue{Path: full, Severity: "error", Message: err.Error()})
			continue
		}
		matches := h1Re.FindAll(b, -1)
		if len(matches) == 0 {
			issues = append(issues, Issue{Path: full, Severity: "warning", Message: label + " has no H1 heading"})
		} else if len(matches) > 1 {
			issues = append(issues, Issue{Path: full, Severity: "warning", Message: fmt.Sprintf("%s has %d H1 headings; expected 1", label, len(matches))})
		}
	}
	return issues, nil
}

// secretRe is a deliberately conservative scanner for obvious credential
// shapes: classic GitHub PATs, AWS access key IDs, generic high-entropy
// patterns labeled SECRET/TOKEN. It is not a substitute for a real secret
// scanner.
var secretRe = regexp.MustCompile(`(?i)(ghp_[A-Za-z0-9]{30,}|AKIA[0-9A-Z]{12,}|secret[_-]?key\s*[:=]\s*['"][^'"]{12,}['"]|api[_-]?token\s*[:=]\s*['"][^'"]{12,}['"])`)

func validateNoSecrets(root string) ([]Issue, error) {
	var issues []Issue
	err := filepath.WalkDir(repo.SharedDir(root), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		if secretRe.Match(b) {
			issues = append(issues, Issue{Path: p, Severity: "error", Message: "possible secret/credential token detected"})
		}
		return nil
	})
	return issues, err
}

// splitFrontmatter peels a leading YAML frontmatter block (delimited by
// `---` lines) off `b`. Returns (frontmatter, body, error). If no
// frontmatter is found, returns (nil, b, nil).
func splitFrontmatter(b []byte) ([]byte, []byte, error) {
	if !bytes.HasPrefix(b, []byte("---\n")) && !bytes.HasPrefix(b, []byte("---\r\n")) {
		return nil, b, nil
	}
	rest := b[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return nil, b, errors.New("frontmatter: opening `---` without closing `---`")
	}
	fm := rest[:end]
	tail := rest[end+len("\n---"):]
	tail = bytes.TrimLeft(tail, "\r\n")
	return fm, tail, nil
}

func writeFromTemplate(dst, tmplPath string, data any) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("refusing to overwrite existing file: %s", dst)
	}
	raw, err := templates.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	t, err := template.New(filepath.Base(tmplPath)).Parse(string(raw))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return t.Execute(f, data)
}

var nameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func validateName(name string) error {
	if len(name) < 2 || len(name) > 64 || !nameRe.MatchString(name) {
		return fmt.Errorf("invalid name %q: use lowercase letters, digits, and `-` (2-64 chars, no leading/trailing dash)", name)
	}
	return nil
}

func humanize(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}
