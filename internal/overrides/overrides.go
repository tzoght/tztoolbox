// Package overrides reads and writes overrides/<tool>/disabled.yaml so
// `tzcli enable` and `tzcli disable` can toggle inclusion per artifact.
package overrides

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/repo"
)

// Disabled is the on-disk shape of overrides/<tool>/disabled.yaml.
type Disabled struct {
	Commands []string `yaml:"commands,omitempty"`
	Rules    []string `yaml:"rules,omitempty"`
	Skills   []string `yaml:"skills,omitempty"`
}

// Kind is "command" | "rule" | "skill" — the artifact category.
type Kind string

const (
	KindCommand Kind = "command"
	KindRule    Kind = "rule"
	KindSkill   Kind = "skill"
)

// ParseKind validates the string form.
func ParseKind(s string) (Kind, error) {
	switch Kind(s) {
	case KindCommand, KindRule, KindSkill:
		return Kind(s), nil
	}
	return "", fmt.Errorf("unknown kind %q (expected command|rule|skill)", s)
}

// Path returns overrides/<tool>/disabled.yaml.
func Path(root string, t model.Tool) string {
	return filepath.Join(repo.OverridesDir(root), t.String(), "disabled.yaml")
}

// Load reads disabled.yaml; returns empty struct if absent.
func Load(root string, t model.Tool) (Disabled, error) {
	var d Disabled
	b, err := os.ReadFile(Path(root, t))
	if errors.Is(err, fs.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if err := yaml.Unmarshal(b, &d); err != nil {
		return d, err
	}
	return d, nil
}

// Save writes disabled.yaml. If the resulting struct is empty (no entries),
// the file is removed instead.
func Save(root string, t model.Tool, d Disabled) error {
	p := Path(root, t)
	if len(d.Commands)+len(d.Rules)+len(d.Skills) == 0 {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// Disable adds `name` to the disabled list for kind/tool. Idempotent.
func Disable(root string, t model.Tool, k Kind, name string) error {
	d, err := Load(root, t)
	if err != nil {
		return err
	}
	switch k {
	case KindCommand:
		d.Commands = addUnique(d.Commands, name)
	case KindRule:
		d.Rules = addUnique(d.Rules, name)
	case KindSkill:
		d.Skills = addUnique(d.Skills, name)
	}
	return Save(root, t, d)
}

// Enable removes `name` from the disabled list for kind/tool. Idempotent.
func Enable(root string, t model.Tool, k Kind, name string) error {
	d, err := Load(root, t)
	if err != nil {
		return err
	}
	switch k {
	case KindCommand:
		d.Commands = remove(d.Commands, name)
	case KindRule:
		d.Rules = remove(d.Rules, name)
	case KindSkill:
		d.Skills = remove(d.Skills, name)
	}
	return Save(root, t, d)
}

func addUnique(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	xs = append(xs, s)
	sort.Strings(xs)
	return xs
}

func remove(xs []string, s string) []string {
	out := xs[:0]
	for _, x := range xs {
		if x != s {
			out = append(out, x)
		}
	}
	return out
}
