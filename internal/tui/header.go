package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tzoght/tztoolbox/internal/doctor"
	"github.com/tzoght/tztoolbox/internal/model"
)

// refreshHeader recomputes the status bar information by running doctor.Run
// in the background and posting a [headerRefreshedMsg] when it returns.
//
// We call this on init and after every dispatch; doctor.Run does some
// shelling out so it's not free, but each call is bounded to a few ms and
// the user benefits from up-to-date drift counts.
func (m Model) refreshHeader() tea.Cmd {
	repoRoot := m.repoRoot
	version := m.header.Version
	return func() tea.Msg {
		info := HeaderInfo{Version: version, RepoRoot: repoRoot}
		if repoRoot == "" {
			return headerRefreshedMsg{info: info}
		}
		rep, err := doctor.Run(repoRoot)
		if err != nil {
			return headerRefreshedMsg{info: info}
		}
		info.Branch = rep.Repo.Branch
		info.DriftCount = len(rep.Drift.Entries)

		var parts []string
		for _, t := range model.AllTools {
			if rep.Tools[t].Detected {
				parts = append(parts, t.String()+"*")
			} else {
				parts = append(parts, t.String())
			}
		}
		info.ToolsLine = strings.Join(parts, " ")
		return headerRefreshedMsg{info: info}
	}
}

// refreshPalette walks the live Cobra tree and replaces the static palette
// items with the actual subcommands and their short help, prefixed by the
// mnemonic key when one is registered.
func (m Model) refreshPalette() tea.Cmd {
	factory := m.factory
	return func() tea.Msg {
		if factory == nil {
			return paletteRefreshedMsg{items: builtinPaletteItems()}
		}
		root := factory()
		if root == nil {
			return paletteRefreshedMsg{items: builtinPaletteItems()}
		}
		// Map subcommand name -> mnemonic key so we can decorate items.
		keyByCmd := map[string]string{}
		for _, mn := range mnemonics {
			keyByCmd[mn.Cmd] = mn.Key
		}
		var items []list.Item
		for _, c := range root.Commands() {
			if c.Hidden || c.Name() == "completion" {
				continue
			}
			desc := c.Short
			if k, ok := keyByCmd[c.Name()]; ok {
				desc = "[" + k + "]  " + desc
			}
			items = append(items, paletteItem{command: c.Name(), desc: desc})
		}
		items = append(items, paletteItem{command: "quit", desc: "[q]  Exit the shell"})
		return paletteRefreshedMsg{items: items}
	}
}
