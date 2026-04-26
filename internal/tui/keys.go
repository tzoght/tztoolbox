package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap names every key binding the TUI honors. Used both for the Update
// switch and for the bubbles/help overlay.
type keyMap struct {
	Submit       key.Binding
	Quit         key.Binding
	Cancel       key.Binding
	ClearScreen  key.Binding
	HistoryUp    key.Binding
	HistoryDown  key.Binding
	Complete     key.Binding
	Palette      key.Binding
	PaletteClose key.Binding
	Help         key.Binding
	Menu         key.Binding
	ScrollUp     key.Binding
	ScrollDown   key.Binding
	ScrollHome   key.Binding
	ScrollEnd    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "quit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel/clear input"),
		),
		ClearScreen: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear output"),
		),
		HistoryUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "previous"),
		),
		HistoryDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "next"),
		),
		Complete: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "complete"),
		),
		Palette: key.NewBinding(
			key.WithKeys("ctrl+k"),
			key.WithHelp("ctrl+k", "palette"),
		),
		PaletteClose: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Menu: key.NewBinding(
			// F1 is the canonical "show me the menu" key; ctrl+g is a
			// terminal-friendly fallback ("guide") for keyboards or
			// terminal emulators that swallow F1.
			key.WithKeys("f1", "ctrl+g"),
			key.WithHelp("f1/ctrl+g", "menu"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "scroll down"),
		),
		ScrollHome: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "top"),
		),
		ScrollEnd: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "bottom"),
		),
	}
}

// ShortHelp implements help.KeyMap for the footer hint line.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Complete, k.Menu, k.Palette, k.HistoryUp, k.ClearScreen, k.Help, k.Quit}
}

// FullHelp implements help.KeyMap for the expanded help overlay.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Cancel, k.Quit},
		{k.HistoryUp, k.HistoryDown, k.Complete, k.Palette, k.Menu},
		{k.ScrollUp, k.ScrollDown, k.ScrollHome, k.ScrollEnd},
		{k.ClearScreen, k.PaletteClose, k.Help},
	}
}
