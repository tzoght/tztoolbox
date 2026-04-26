// Package model defines the core types used across tzcli: the supported tools
// (Cursor, Claude Code, OpenAI Codex CLI) and the per-tool layout conventions.
package model

import (
	"fmt"
	"path/filepath"
)

// Tool identifies one of the supported AI CLIs that tztoolbox can target.
type Tool string

const (
	Cursor Tool = "cursor"
	Claude Tool = "claude"
	Codex  Tool = "codex"
)

// AllTools is the canonical, deterministic order tzcli iterates targets in.
var AllTools = []Tool{Cursor, Claude, Codex}

// ParseTool returns the Tool for the given string, or an error if unknown.
func ParseTool(s string) (Tool, error) {
	switch Tool(s) {
	case Cursor, Claude, Codex:
		return Tool(s), nil
	default:
		return "", fmt.Errorf("unknown tool %q (expected cursor|claude|codex)", s)
	}
}

// String returns the lowercase tool identifier.
func (t Tool) String() string { return string(t) }

// RepoSubdir returns the in-repo native directory for this tool, e.g. ".cursor".
func (t Tool) RepoSubdir() string { return "." + string(t) }

// HomeSubdir returns the user-home native directory for this tool, e.g. ".cursor".
func (t Tool) HomeSubdir() string { return "." + string(t) }

// CommandsSubdir returns the per-tool subfolder where slash commands live.
// Codex calls them "prompts"; the others use "commands".
func (t Tool) CommandsSubdir() string {
	if t == Codex {
		return "prompts"
	}
	return "commands"
}

// SkillsSubdir returns the per-tool skills subfolder. All three use "skills".
func (t Tool) SkillsSubdir() string { return "skills" }

// RulesMode describes how a tool consumes always-on guidance ("rules").
type RulesMode int

const (
	// RulesAsFiles means each rule is its own file under <tool>/rules/.
	RulesAsFiles RulesMode = iota
	// RulesConcatenated means all rules are concatenated into a single file
	// (e.g. ~/.claude/CLAUDE.md, ~/.codex/AGENTS.md).
	RulesConcatenated
)

// RulesMode reports how this tool consumes rules.
func (t Tool) RulesMode() RulesMode {
	if t == Cursor {
		return RulesAsFiles
	}
	return RulesConcatenated
}

// RulesPath returns the path *relative* to the tool's native root where rules
// should be written. For RulesAsFiles, this is a directory ("rules"). For
// RulesConcatenated, it is the single file ("CLAUDE.md", "AGENTS.md").
func (t Tool) RulesPath() string {
	switch t {
	case Cursor:
		return "rules"
	case Claude:
		return "CLAUDE.md"
	case Codex:
		return "AGENTS.md"
	default:
		return ""
	}
}

// CommandFileExt returns the on-disk extension to use when writing commands
// for this tool. All three use ".md" today, but kept here for future drift.
func (t Tool) CommandFileExt() string { return ".md" }

// RuleFileExt returns the on-disk extension for an individual rule file.
// Cursor uses ".mdc" (with frontmatter); the concatenated tools never write
// individual rule files, so this returns "" for them.
func (t Tool) RuleFileExt() string {
	if t == Cursor {
		return ".mdc"
	}
	return ""
}

// HomeDir returns the absolute path to ~/<tool>'s home (e.g. ~/.cursor).
func (t Tool) HomeDir(homeDir string) string {
	return filepath.Join(homeDir, t.HomeSubdir())
}
