// Package inventory enumerates what's actually deployed under each tool's
// home directory (~/.cursor, ~/.claude, ~/.codex), side-by-side with the
// in-repo native trees (<repo>/.<tool>). It powers `tzcli installed`.
//
// Inventory differs from the diffsearch and render packages in two ways:
//
//  1. It groups the filesystem at the *artifact* level (one entry per
//     command/rule/skill) rather than reporting raw byte-level drift.
//  2. It correlates each artifact with the install manifest written by the
//     installer, so callers can answer "did tztoolbox put this here?".
package inventory

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tzoght/tztoolbox/internal/installer"
	"github.com/tzoght/tztoolbox/internal/model"
)

// Kind names the artifact category for an inventory entry.
type Kind string

const (
	KindCommand Kind = "command"
	KindRule    Kind = "rule"
	KindSkill   Kind = "skill"
)

// AllKinds is the canonical iteration order.
var AllKinds = []Kind{KindCommand, KindRule, KindSkill}

// ParseKind validates a user-supplied kind string.
func ParseKind(s string) (Kind, error) {
	switch Kind(strings.ToLower(s)) {
	case KindCommand, KindRule, KindSkill:
		return Kind(strings.ToLower(s)), nil
	}
	return "", fmt.Errorf("unknown kind %q (expected command|rule|skill)", s)
}

// Status is how an artifact compares between repo and home.
type Status string

const (
	StatusInSync     Status = "in_sync"
	StatusOnlyInRepo Status = "only_in_repo"
	StatusOnlyInHome Status = "only_in_home"
	StatusChanged    Status = "changed"
)

// Item is a single inventory entry.
type Item struct {
	Tool    model.Tool `json:"tool"`
	Kind    Kind       `json:"kind"`
	Name    string     `json:"name"`
	RelPath string     `json:"rel_path"` // path relative to the tool dir; "" for skill dirs
	Repo    bool       `json:"repo"`
	Home    bool       `json:"home"`
	Managed bool       `json:"managed"`
	Status  Status     `json:"status"`
}

// ToolEntry bundles a tool's inventory rows with the paths they came from.
type ToolEntry struct {
	Tool         model.Tool `json:"tool"`
	RepoDir      string     `json:"repo_dir"`
	HomeDir      string     `json:"home_dir"`
	ManifestPath string     `json:"manifest_path"`
	HomeExists   bool       `json:"home_exists"`
	Items        []Item     `json:"items"`
}

// Report is the full output of Scan, indexed by tool.
type Report struct {
	Tools map[model.Tool]ToolEntry `json:"tools"`
}

// Options configures a Scan call.
type Options struct {
	// ManagedOnly hides any entry whose path is not listed in the home
	// directory's .tztoolbox.manifest. Useful for "what does tztoolbox own
	// here?" rather than "what's in this directory?".
	ManagedOnly bool
	// Kinds restricts to a subset; empty = all.
	Kinds []Kind
}

// Scan walks <root>/.<tool> and <home>/.<tool> for each requested tool and
// returns a Report. Both root and home may contain dirs that don't exist;
// missing locations are treated as empty. A nil/empty `tools` slice scans
// all supported tools in the canonical order.
func Scan(root, home string, tools []model.Tool, opts Options) (Report, error) {
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return Report{}, fmt.Errorf("home dir: %w", err)
		}
		home = h
	}
	if len(tools) == 0 {
		tools = model.AllTools
	}

	rep := Report{Tools: map[model.Tool]ToolEntry{}}
	for _, t := range tools {
		entry, err := scanTool(root, home, t, opts)
		if err != nil {
			return rep, fmt.Errorf("scan %s: %w", t, err)
		}
		rep.Tools[t] = entry
	}
	return rep, nil
}

func scanTool(root, home string, t model.Tool, opts Options) (ToolEntry, error) {
	entry := ToolEntry{
		Tool:         t,
		RepoDir:      filepath.Join(root, t.RepoSubdir()),
		HomeDir:      filepath.Join(home, t.HomeSubdir()),
		ManifestPath: filepath.Join(home, t.HomeSubdir(), installer.ManifestName),
	}
	if _, err := os.Stat(entry.HomeDir); err == nil {
		entry.HomeExists = true
	}

	managed, err := installer.ReadManifest(entry.ManifestPath)
	if err != nil {
		return entry, fmt.Errorf("read manifest: %w", err)
	}
	managedSet := make(map[string]struct{}, len(managed))
	for _, m := range managed {
		managedSet[m] = struct{}{}
	}

	var items []Item
	for _, k := range AllKinds {
		if !wantKind(k, opts.Kinds) {
			continue
		}
		ki, err := scanKind(entry, t, k, managedSet)
		if err != nil {
			return entry, err
		}
		items = append(items, ki...)
	}

	if opts.ManagedOnly {
		filtered := items[:0]
		for _, it := range items {
			if it.Managed {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Name < items[j].Name
	})
	entry.Items = items
	return entry, nil
}

func wantKind(k Kind, kinds []Kind) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, x := range kinds {
		if x == k {
			return true
		}
	}
	return false
}

func scanKind(e ToolEntry, t model.Tool, k Kind, managed map[string]struct{}) ([]Item, error) {
	switch k {
	case KindCommand:
		return scanCommands(e, t, managed)
	case KindRule:
		return scanRules(e, t, managed)
	case KindSkill:
		return scanSkills(e, t, managed)
	}
	return nil, nil
}

// scanCommands handles cursor/claude `commands/*.md` and codex `prompts/*.md`.
func scanCommands(e ToolEntry, t model.Tool, managed map[string]struct{}) ([]Item, error) {
	sub := t.CommandsSubdir()
	repoDir := filepath.Join(e.RepoDir, sub)
	homeDir := filepath.Join(e.HomeDir, sub)

	repoFiles, err := listMarkdown(repoDir)
	if err != nil {
		return nil, err
	}
	homeFiles, err := listMarkdown(homeDir)
	if err != nil {
		return nil, err
	}

	names := unionKeys(repoFiles, homeFiles)
	out := make([]Item, 0, len(names))
	for _, name := range names {
		rel := filepath.ToSlash(filepath.Join(sub, name+".md"))
		it := Item{
			Tool:    t,
			Kind:    KindCommand,
			Name:    name,
			RelPath: rel,
			Repo:    repoFiles[name] != nil,
			Home:    homeFiles[name] != nil,
		}
		_, it.Managed = managed[rel]
		it.Status = fileStatus(repoFiles[name], homeFiles[name], it.Repo, it.Home)
		out = append(out, it)
	}
	return out, nil
}

// scanRules handles two shapes: cursor's per-file `.mdc` rules, and the
// concatenated `CLAUDE.md` / `AGENTS.md` bundles for claude/codex.
func scanRules(e ToolEntry, t model.Tool, managed map[string]struct{}) ([]Item, error) {
	switch t.RulesMode() {
	case model.RulesAsFiles:
		dir := t.RulesPath() // "rules"
		repoDir := filepath.Join(e.RepoDir, dir)
		homeDir := filepath.Join(e.HomeDir, dir)
		ext := t.RuleFileExt() // ".mdc"

		repoFiles, err := listByExt(repoDir, ext)
		if err != nil {
			return nil, err
		}
		homeFiles, err := listByExt(homeDir, ext)
		if err != nil {
			return nil, err
		}

		names := unionKeys(repoFiles, homeFiles)
		out := make([]Item, 0, len(names))
		for _, name := range names {
			rel := filepath.ToSlash(filepath.Join(dir, name+ext))
			it := Item{
				Tool:    t,
				Kind:    KindRule,
				Name:    name,
				RelPath: rel,
				Repo:    repoFiles[name] != nil,
				Home:    homeFiles[name] != nil,
			}
			_, it.Managed = managed[rel]
			it.Status = fileStatus(repoFiles[name], homeFiles[name], it.Repo, it.Home)
			out = append(out, it)
		}
		return out, nil

	case model.RulesConcatenated:
		// Single bundle file (CLAUDE.md or AGENTS.md). Treat the whole file
		// as one rule-bundle artifact named after the file.
		rel := t.RulesPath()
		repoPath := filepath.Join(e.RepoDir, rel)
		homePath := filepath.Join(e.HomeDir, rel)

		repoBytes, repoOK, err := readIfExists(repoPath)
		if err != nil {
			return nil, err
		}
		homeBytes, homeOK, err := readIfExists(homePath)
		if err != nil {
			return nil, err
		}
		if !repoOK && !homeOK {
			return nil, nil
		}

		var bRepo, bHome []byte
		if repoOK {
			bRepo = repoBytes
		}
		if homeOK {
			bHome = homeBytes
		}

		it := Item{
			Tool:    t,
			Kind:    KindRule,
			Name:    rel, // e.g. "CLAUDE.md"
			RelPath: rel,
			Repo:    repoOK,
			Home:    homeOK,
			Status:  fileStatus(bRepo, bHome, repoOK, homeOK),
		}
		_, it.Managed = managed[rel]
		return []Item{it}, nil
	}
	return nil, nil
}

// scanSkills enumerates skill *directories* under <tool>/skills/, comparing
// the entire subtree of each. Any byte-level drift inside flips the skill
// to StatusChanged.
func scanSkills(e ToolEntry, t model.Tool, managed map[string]struct{}) ([]Item, error) {
	sub := t.SkillsSubdir()
	repoDir := filepath.Join(e.RepoDir, sub)
	homeDir := filepath.Join(e.HomeDir, sub)

	repoSkills, err := listSubdirs(repoDir)
	if err != nil {
		return nil, err
	}
	homeSkills, err := listSubdirs(homeDir)
	if err != nil {
		return nil, err
	}

	names := unionStringSets(repoSkills, homeSkills)
	out := make([]Item, 0, len(names))
	for _, name := range names {
		repoPath := filepath.Join(repoDir, name)
		homePath := filepath.Join(homeDir, name)

		inRepo := repoSkills[name]
		inHome := homeSkills[name]

		it := Item{
			Tool:    t,
			Kind:    KindSkill,
			Name:    name,
			RelPath: filepath.ToSlash(filepath.Join(sub, name)),
			Repo:    inRepo,
			Home:    inHome,
		}

		// Skills are managed if *any* of their files appear in the manifest.
		// The installer writes one manifest entry per file inside the skill
		// directory, so checking for a prefix match is the precise test.
		prefix := it.RelPath + "/"
		for m := range managed {
			if strings.HasPrefix(m, prefix) {
				it.Managed = true
				break
			}
		}

		switch {
		case inRepo && !inHome:
			it.Status = StatusOnlyInRepo
		case !inRepo && inHome:
			it.Status = StatusOnlyInHome
		default:
			same, err := dirsEqual(repoPath, homePath)
			if err != nil {
				return nil, err
			}
			if same {
				it.Status = StatusInSync
			} else {
				it.Status = StatusChanged
			}
		}
		out = append(out, it)
	}
	return out, nil
}

// fileStatus computes Status from the (possibly nil) byte contents of a file
// in the repo and home trees.
func fileStatus(repo, home []byte, inRepo, inHome bool) Status {
	switch {
	case inRepo && !inHome:
		return StatusOnlyInRepo
	case !inRepo && inHome:
		return StatusOnlyInHome
	case bytes.Equal(repo, home):
		return StatusInSync
	default:
		return StatusChanged
	}
}

// listMarkdown reads `dir` and returns a map of basename -> file bytes for
// every `.md` file. Missing dirs return an empty map (not an error).
func listMarkdown(dir string) (map[string][]byte, error) {
	return listByExt(dir, ".md")
}

func listByExt(dir, ext string) (map[string][]byte, error) {
	out := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		out[strings.TrimSuffix(e.Name(), ext)] = b
	}
	return out, nil
}

// listSubdirs returns a set of subdirectory names directly under `dir`.
func listSubdirs(dir string) (map[string]bool, error) {
	out := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = true
		}
	}
	return out, nil
}

// dirsEqual reports whether two directory trees are byte-equal across every
// regular file. Missing dirs are not equal to existing ones unless both are
// missing.
func dirsEqual(a, b string) (bool, error) {
	aFiles, err := walkFiles(a)
	if err != nil {
		return false, err
	}
	bFiles, err := walkFiles(b)
	if err != nil {
		return false, err
	}
	if len(aFiles) != len(bFiles) {
		return false, nil
	}
	for path, ab := range aFiles {
		bb, ok := bFiles[path]
		if !ok || !bytes.Equal(ab, bb) {
			return false, nil
		}
	}
	return true, nil
}

func walkFiles(root string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		out[filepath.ToSlash(rel)] = b
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	return out, err
}

func readIfExists(path string) ([]byte, bool, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func unionKeys(a, b map[string][]byte) []string {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func unionStringSets(a, b map[string]bool) []string {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
