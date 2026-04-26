package render

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/tzoght/tztoolbox/internal/model"
)

// DriftKind identifies how a single path differs between the live in-repo
// native tree and the freshly-rendered expected tree.
type DriftKind string

const (
	// DriftMissing means the path exists in the rendered output but not on disk.
	DriftMissing DriftKind = "missing"
	// DriftExtra means the path exists on disk but not in the rendered output.
	DriftExtra DriftKind = "extra"
	// DriftChanged means the file exists in both trees but content differs.
	DriftChanged DriftKind = "changed"
)

// DriftEntry describes a single divergence.
type DriftEntry struct {
	Tool model.Tool
	Path string // relative to <root>/.<tool>
	Kind DriftKind
}

// DriftReport is the full output of CheckDrift.
type DriftReport struct {
	Entries []DriftEntry
}

// HasDrift reports whether any divergence was detected.
func (r DriftReport) HasDrift() bool { return len(r.Entries) > 0 }

// CheckDrift renders the canonical sources into a temp directory and compares
// the result, file by file, against the live in-repo native trees. Drift
// indicates `tzcli sync` would produce different output than what's checked
// in.
func CheckDrift(root string, tools []model.Tool) (DriftReport, error) {
	if len(tools) == 0 {
		tools = model.AllTools
	}

	tmp, err := os.MkdirTemp("", "tzcli-drift-*")
	if err != nil {
		return DriftReport{}, fmt.Errorf("mktemp: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	// Mirror the layout: copy shared/ + overrides/ into tmp, then render
	// into tmp so the tool dirs land there.
	for _, sub := range []string{"shared", "overrides"} {
		src := filepath.Join(root, sub)
		if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err := copyDir(src, filepath.Join(tmp, sub)); err != nil {
			return DriftReport{}, fmt.Errorf("seed %s: %w", sub, err)
		}
	}
	if _, err := Render(tmp, Options{Tools: tools}); err != nil {
		return DriftReport{}, fmt.Errorf("render to tmp: %w", err)
	}

	var report DriftReport
	for _, t := range tools {
		entries, err := diffTrees(filepath.Join(tmp, t.RepoSubdir()), filepath.Join(root, t.RepoSubdir()), t)
		if err != nil {
			return report, err
		}
		report.Entries = append(report.Entries, entries...)
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		if report.Entries[i].Tool != report.Entries[j].Tool {
			return report.Entries[i].Tool < report.Entries[j].Tool
		}
		return report.Entries[i].Path < report.Entries[j].Path
	})
	return report, nil
}

// diffTrees walks expected and actual side-by-side, returning DriftEntry per
// differing path. Both directories may not exist; absence equals "empty".
func diffTrees(expected, actual string, tool model.Tool) ([]DriftEntry, error) {
	expectedFiles, err := listFiles(expected)
	if err != nil {
		return nil, err
	}
	actualFiles, err := listFiles(actual)
	if err != nil {
		return nil, err
	}

	var out []DriftEntry
	for path, expBytes := range expectedFiles {
		actBytes, ok := actualFiles[path]
		if !ok {
			out = append(out, DriftEntry{Tool: tool, Path: path, Kind: DriftMissing})
			continue
		}
		if !bytes.Equal(expBytes, actBytes) {
			out = append(out, DriftEntry{Tool: tool, Path: path, Kind: DriftChanged})
		}
	}
	for path := range actualFiles {
		if _, ok := expectedFiles[path]; !ok {
			out = append(out, DriftEntry{Tool: tool, Path: path, Kind: DriftExtra})
		}
	}
	return out, nil
}

// listFiles returns a map of <relpath> -> file contents for every regular
// file beneath dir. A missing dir is treated as an empty result.
func listFiles(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
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
