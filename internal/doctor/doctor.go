// Package doctor inspects the local environment and reports on tztoolbox
// readiness: which tool homes exist, what versions are installed, what
// would be deployed to each, whether there is drift between shared/ and
// the in-repo native trees, and the state of related tooling (gh, op).
package doctor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/render"
	"github.com/tzoght/tztoolbox/internal/repo"
)

// Report is the full doctor output. It's also marshallable to JSON.
type Report struct {
	Repo    RepoInfo                  `json:"repo"`
	Tools   map[model.Tool]ToolStatus `json:"tools"`
	Drift   render.DriftReport        `json:"drift"`
	Helpers map[string]HelperStatus   `json:"helpers"`
	Counts  map[string]map[string]int `json:"counts"` // tool -> kind -> count
}

// RepoInfo describes the tztoolbox checkout itself.
type RepoInfo struct {
	Root          string `json:"root"`
	Branch        string `json:"branch"`
	OriginURL     string `json:"origin_url"`
	IsDirty       bool   `json:"is_dirty"`
	HasGitDir     bool   `json:"has_git_dir"`
	DefaultBranch string `json:"default_branch"`
}

// ToolStatus describes one tool's readiness.
type ToolStatus struct {
	Detected   bool   `json:"detected"`
	HomePath   string `json:"home_path"`
	HomeExists bool   `json:"home_exists"`
	Version    string `json:"version,omitempty"`
}

// HelperStatus is for adjacent tools (gh, op, etc).
type HelperStatus struct {
	Found   bool   `json:"found"`
	Version string `json:"version,omitempty"`
}

// Run gathers the report. Network calls are best-effort; failures are
// captured in fields rather than returning an error.
func Run(root string) (Report, error) {
	rep := Report{
		Repo:    repoInfo(root),
		Tools:   map[model.Tool]ToolStatus{},
		Helpers: map[string]HelperStatus{},
		Counts:  map[string]map[string]int{},
	}

	home, _ := os.UserHomeDir()
	for _, t := range model.AllTools {
		rep.Tools[t] = inspectTool(t, home)
		rep.Counts[t.String()] = countArtifacts(root, t)
	}

	rep.Helpers["gh"] = lookupHelper("gh", "--version")
	rep.Helpers["op"] = lookupHelper("op", "--version")
	rep.Helpers["go"] = lookupHelper("go", "version")
	rep.Helpers["git"] = lookupHelper("git", "--version")
	// cursor-agent is the Cursor CLI binary (`agent` / `cursor-agent`).
	// When present, the same shared/ commands and skills tzcli renders
	// into ~/.cursor/ are usable from a terminal via `agent` outside the IDE.
	rep.Helpers["cursor-agent"] = lookupHelper("cursor-agent", "--version")

	drift, err := render.CheckDrift(root, model.AllTools)
	if err == nil {
		rep.Drift = drift
	}
	return rep, nil
}

func repoInfo(root string) RepoInfo {
	info := RepoInfo{Root: root}
	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		info.HasGitDir = true
	}
	if out, err := runIn(root, "git", "branch", "--show-current"); err == nil {
		info.Branch = strings.TrimSpace(out)
	}
	if out, err := runIn(root, "git", "config", "--get", "remote.origin.url"); err == nil {
		info.OriginURL = strings.TrimSpace(out)
	}
	if out, err := runIn(root, "git", "status", "--porcelain"); err == nil {
		info.IsDirty = strings.TrimSpace(out) != ""
	}
	if out, err := runIn(root, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		s := strings.TrimSpace(out)
		s = strings.TrimPrefix(s, "origin/")
		info.DefaultBranch = s
	}
	return info
}

func inspectTool(t model.Tool, home string) ToolStatus {
	hp := t.HomeDir(home)
	st := ToolStatus{HomePath: hp}
	if _, err := os.Stat(hp); err == nil {
		st.HomeExists = true
		st.Detected = true
	}
	switch t {
	case model.Cursor:
		if v, ok := lookupVersion("cursor", "--version"); ok {
			st.Version = v
			st.Detected = true
		}
	case model.Claude:
		if v, ok := lookupVersion("claude", "--version"); ok {
			st.Version = v
			st.Detected = true
		}
	case model.Codex:
		if v, ok := lookupVersion("codex", "--version"); ok {
			st.Version = v
			st.Detected = true
		}
	}
	return st
}

func countArtifacts(root string, t model.Tool) map[string]int {
	counts := map[string]int{"commands": 0, "rules": 0, "skills": 0}
	if entries, err := os.ReadDir(repo.SharedCommandsDir(root)); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				counts["commands"]++
			}
		}
	}
	if entries, err := os.ReadDir(repo.SharedRulesDir(root)); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				counts["rules"]++
			}
		}
	}
	if entries, err := os.ReadDir(repo.SharedSkillsDir(root)); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				counts["skills"]++
			}
		}
	}
	return counts
}

func lookupHelper(bin string, args ...string) HelperStatus {
	if _, err := exec.LookPath(bin); err != nil {
		return HelperStatus{Found: false}
	}
	if v, ok := lookupVersion(bin, args...); ok {
		return HelperStatus{Found: true, Version: v}
	}
	return HelperStatus{Found: true}
}

func lookupVersion(bin string, args ...string) (string, bool) {
	if _, err := exec.LookPath(bin); err != nil {
		return "", false
	}
	out, err := exec.Command(bin, args...).CombinedOutput() //nolint:gosec // user-controlled bin name comes from a fixed allowlist
	if err != nil {
		return "", false
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	return line, true
}

func runIn(dir string, bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...) //nolint:gosec // bin is hardcoded to git
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("%s %v: %w", bin, args, ee)
		}
		if errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		return "", err
	}
	return string(out), nil
}
