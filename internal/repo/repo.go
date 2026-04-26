// Package repo finds the tztoolbox repository root from any subdirectory.
//
// Discovery walks upward from the start directory looking for a marker that
// uniquely identifies a tztoolbox checkout. The marker is a `go.mod` whose
// module path is `github.com/tzoght/tztoolbox` (the canonical module path of
// this repo). A `.tztoolbox` sentinel file is also accepted so users can
// shadow the lookup in non-Go contexts.
package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModulePath is the Go module path used to identify a tztoolbox repo root.
const ModulePath = "github.com/tzoght/tztoolbox"

// ErrNotFound indicates Discover walked all the way to the filesystem root
// without finding a tztoolbox checkout marker.
var ErrNotFound = errors.New("tztoolbox repo root not found (no go.mod with module " + ModulePath + " and no .tztoolbox sentinel)")

// Discover returns the absolute path to the tztoolbox repository root by
// walking upward from start. If start is empty, the current working
// directory is used.
func Discover(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		start = wd
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", start, err)
	}

	dir := abs
	for {
		if isRepoRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// isRepoRoot reports whether dir is a tztoolbox repository root.
func isRepoRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".tztoolbox")); err == nil {
		return true
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(mod), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return path == ModulePath
		}
	}
	return false
}

// SharedDir returns <root>/shared.
func SharedDir(root string) string { return filepath.Join(root, "shared") }

// OverridesDir returns <root>/overrides.
func OverridesDir(root string) string { return filepath.Join(root, "overrides") }

// SharedCommandsDir returns <root>/shared/commands.
func SharedCommandsDir(root string) string { return filepath.Join(root, "shared", "commands") }

// SharedRulesDir returns <root>/shared/rules.
func SharedRulesDir(root string) string { return filepath.Join(root, "shared", "rules") }

// SharedSkillsDir returns <root>/shared/skills.
func SharedSkillsDir(root string) string { return filepath.Join(root, "shared", "skills") }

// CursorOverridesDir returns <root>/overrides/cursor.
func CursorOverridesDir(root string) string { return filepath.Join(root, "overrides", "cursor") }
