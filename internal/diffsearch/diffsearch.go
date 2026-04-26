// Package diffsearch implements `tzcli search` (substring/regex search over
// shared/ content). The drift-detection used by `tzcli sync --check` and the
// status bar lives in internal/render.CheckDrift.
package diffsearch

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tzoght/tztoolbox/internal/repo"
)

// Match describes a single search hit.
type Match struct {
	Path string
	Line int
	Text string
}

// SearchOptions configures Search.
type SearchOptions struct {
	// Kinds restricts the search domain. Allowed values: "commands", "rules",
	// "skills". Empty = all.
	Kinds []string
	// IgnoreCase makes the regex case-insensitive.
	IgnoreCase bool
}

// Search runs a regex search over shared/ content. The pattern is compiled
// once; an invalid pattern returns an error.
func Search(root, pattern string, opts SearchOptions) ([]Match, error) {
	flags := ""
	if opts.IgnoreCase {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return nil, err
	}

	var roots []string
	want := func(k string) bool {
		if len(opts.Kinds) == 0 {
			return true
		}
		for _, kk := range opts.Kinds {
			if kk == k {
				return true
			}
		}
		return false
	}
	if want("commands") {
		roots = append(roots, repo.SharedCommandsDir(root))
	}
	if want("rules") {
		roots = append(roots, repo.SharedRulesDir(root))
	}
	if want("skills") {
		roots = append(roots, repo.SharedSkillsDir(root))
	}

	var hits []Match
	for _, r := range roots {
		err := filepath.WalkDir(r, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			for i, line := range strings.Split(string(b), "\n") {
				if re.MatchString(line) {
					hits = append(hits, Match{Path: p, Line: i + 1, Text: line})
				}
			}
			return nil
		})
		if err != nil {
			return hits, err
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Path != hits[j].Path {
			return hits[i].Path < hits[j].Path
		}
		return hits[i].Line < hits[j].Line
	})
	return hits, nil
}
