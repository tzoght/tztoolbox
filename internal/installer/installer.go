// Package installer copies the in-repo native trees (.cursor/, .claude/,
// .codex/) into the user's home directory equivalents (~/.cursor, ~/.claude,
// ~/.codex). A small manifest file inside each tool home tracks the relative
// paths tztoolbox is responsible for, so re-running `tzcli install` cleanly
// removes artifacts that have been deleted upstream without disturbing
// unrelated user files.
package installer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tzoght/tztoolbox/internal/model"
)

// ManifestName is the on-disk filename used inside each tool home to track
// which paths tztoolbox owns for that tool.
const ManifestName = ".tztoolbox.manifest"

// Options controls one install pass.
type Options struct {
	// Tools is the set of tools to install for. Empty = all detected.
	Tools []model.Tool
	// HomeDir overrides the user home directory; empty means $HOME.
	HomeDir string
	// DryRun reports the work that would happen without writing.
	DryRun bool
	// Prune removes paths from the previous manifest that are no longer in
	// the current source tree. Defaults to true via NewOptions.
	Prune bool
}

// Result summarizes one install pass.
type Result struct {
	PerTool map[model.Tool]ToolResult
}

// ToolResult is per-tool detail.
type ToolResult struct {
	Installed bool
	Copied    int
	Pruned    int
	HomePath  string
	Skipped   string // reason if Installed=false
}

// Install copies <root>/.<tool> into <home>/.<tool> for each requested tool.
//
// A tool is considered "installed on this machine" if its home dir already
// exists. By default, tools whose home dir does not exist are skipped (a
// clear "tool not installed" signal). Pass an explicit Tools list to force
// installation regardless of presence.
func Install(root string, opts Options) (Result, error) {
	home := opts.HomeDir
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return Result{}, fmt.Errorf("home dir: %w", err)
		}
		home = h
	}

	tools := opts.Tools
	autoSelect := false
	if len(tools) == 0 {
		tools = model.AllTools
		autoSelect = true
	}

	res := Result{PerTool: map[model.Tool]ToolResult{}}
	for _, t := range tools {
		out, err := installTool(root, home, t, opts.DryRun, opts.Prune, autoSelect)
		if err != nil {
			return res, fmt.Errorf("install %s: %w", t, err)
		}
		res.PerTool[t] = out
	}
	return res, nil
}

func installTool(root, home string, t model.Tool, dryRun, prune, autoSelect bool) (ToolResult, error) {
	src := filepath.Join(root, t.RepoSubdir())
	dst := filepath.Join(home, t.HomeSubdir())

	out := ToolResult{HomePath: dst}

	if _, err := os.Stat(src); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			out.Skipped = "no rendered tree at " + src + " (run `tzcli sync` first)"
			return out, nil
		}
		return out, err
	}

	if autoSelect {
		// Skip if user has not installed this tool. Detect by presence of the
		// home dir; create it only when explicitly requested via --only.
		if _, err := os.Stat(dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				out.Skipped = t.String() + " not detected (no " + dst + ")"
				return out, nil
			}
			return out, err
		}
	}

	prev, err := readManifest(filepath.Join(dst, ManifestName))
	if err != nil {
		return out, err
	}

	files, err := listFilesRel(src)
	if err != nil {
		return out, err
	}

	if !dryRun {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return out, err
		}
		for _, rel := range files {
			if err := copyFile(filepath.Join(src, rel), filepath.Join(dst, rel)); err != nil {
				return out, fmt.Errorf("copy %s: %w", rel, err)
			}
		}
	}
	out.Copied = len(files)
	out.Installed = true

	if prune {
		toRemove := manifestDiff(prev, files)
		for _, rel := range toRemove {
			full := filepath.Join(dst, rel)
			if !dryRun {
				if err := os.Remove(full); err != nil && !errors.Is(err, fs.ErrNotExist) {
					return out, fmt.Errorf("prune %s: %w", rel, err)
				}
				removeEmptyParents(dst, filepath.Dir(full))
			}
			out.Pruned++
		}
	}

	if !dryRun {
		if err := writeManifest(filepath.Join(dst, ManifestName), files); err != nil {
			return out, err
		}
	}
	return out, nil
}

// listFilesRel returns sorted slash-separated relative paths for every regular
// file beneath dir.
func listFilesRel(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(out)
	return out, err
}

// manifestDiff returns paths in prev that are no longer in current.
func manifestDiff(prev, current []string) []string {
	cur := map[string]struct{}{}
	for _, p := range current {
		cur[p] = struct{}{}
	}
	var out []string
	for _, p := range prev {
		if _, ok := cur[p]; !ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// ReadManifest parses a tztoolbox install manifest at `path`. The file is one
// path per line, with `#` comments and blank lines ignored. A missing file
// returns (nil, nil) so callers can treat "no prior install" the same as "no
// managed paths". This is the canonical parser for the format written by
// writeManifest; other packages (e.g. internal/inventory) call it directly so
// the read and write sides cannot drift.
func ReadManifest(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		lines = append(lines, l)
	}
	return lines, sc.Err()
}

func readManifest(path string) ([]string, error) { return ReadManifest(path) }

func writeManifest(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("# tztoolbox-managed paths (do not edit; rewritten by `tzcli install`)\n")
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// removeEmptyParents walks upward from `start` (which must be inside `root`)
// removing each directory that became empty after a prune. Stops as soon as
// it hits a non-empty directory or `root` itself.
func removeEmptyParents(root, start string) {
	for {
		if start == root || !strings.HasPrefix(start, root+string(os.PathSeparator)) {
			return
		}
		entries, err := os.ReadDir(start)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(start); err != nil {
			return
		}
		start = filepath.Dir(start)
	}
}
