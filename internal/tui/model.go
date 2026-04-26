package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tzoght/tztoolbox/internal/ui"
)

// Model is the Bubbletea model for the tzcli interactive shell. It owns a
// header bar, a scrollable output viewport, an input line, an optional
// command palette, and a footer help line.
type Model struct {
	factory  RootCmdFactory
	repoRoot string

	width, height int

	header HeaderInfo
	// body is a pointer because Bubbletea's Update returns the Model by
	// value; a value-typed strings.Builder would panic with
	// "non-zero Builder copied by value" once it has been written to.
	body    *strings.Builder
	output  viewport.Model
	input   textinput.Model
	spin    spinner.Model
	help    help.Model
	palette list.Model

	keys   keyMap
	chrome *chromeStyles
	theme  *ui.Theme

	historyPath string
	history     []string
	historyIdx  int // -1 = not navigating
	draft       string

	busy        bool
	busyLabel   string
	paletteOpen bool
	helpOpen    bool

	suggestions []string
}

// HeaderInfo bundles the state shown in the status bar at the top of the
// TUI. It's recomputed on launch and after each command via [refreshHeader].
type HeaderInfo struct {
	Version    string
	RepoRoot   string
	Branch     string
	ToolsLine  string // "cursor* claude  codex*"
	DriftCount int
}

// Options configures NewModel/Run. Most callers only need Factory.
type Options struct {
	Factory  RootCmdFactory
	RepoRoot string // for status bar; "" is fine
	Version  string // displayed in status bar
}

// NewModel constructs a Model with sensible defaults; the model still needs
// to be fed a tea.WindowSizeMsg before its first View().
func NewModel(opts Options) Model {
	chrome := defaultChromeStyles()

	in := textinput.New()
	in.Prompt = ""
	in.Placeholder = "type a command, mnemonic letter (s/d/?/m), or press F1 for the menu"
	in.CharLimit = 1024
	in.Focus()

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = chrome.Spinner

	pl := list.New(nil, paletteDelegate{chrome: chrome}, 40, 12)
	pl.Title = "command palette"
	pl.SetShowStatusBar(false)
	pl.SetFilteringEnabled(true)
	pl.SetShowHelp(false)
	pl.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#1f3a5f", Dark: "#7dd3fc"})

	hp := historyFilePath()
	hist, _ := loadHistory(hp)

	m := Model{
		factory:     opts.Factory,
		repoRoot:    opts.RepoRoot,
		header:      HeaderInfo{Version: opts.Version, RepoRoot: opts.RepoRoot},
		body:        &strings.Builder{},
		output:      vp,
		input:       in,
		spin:        sp,
		help:        help.New(),
		palette:     pl,
		keys:        defaultKeyMap(),
		chrome:      chrome,
		theme:       ui.DefaultTheme(),
		historyPath: hp,
		history:     hist,
		historyIdx:  -1,
	}
	m.help.ShowAll = false

	// Seed palette with builtins immediately so it's usable on Ctrl-K even
	// before the first window-size message; populateCommandPalette refreshes
	// it from the actual factory once Init runs.
	m.palette.SetItems(builtinPaletteItems())
	return m
}

// Init kicks off the spinner ticks, runs an initial header refresh, and
// populates the command palette from the live Cobra tree.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spin.Tick,
		m.refreshHeader(),
		m.refreshPalette(),
		welcomeBanner(m.chrome, m.header.Version),
	)
}

// View renders the four-row layout: header, output, prompt, footer/help.
// Palette overlays the prompt+footer when open.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	// Compose output area + palette/input.
	if m.paletteOpen {
		// Palette takes the bottom slot, replacing input + footer.
		body := m.output.View()
		palette := m.chrome.Palette.Render(m.palette.View())
		return lipgloss.JoinVertical(lipgloss.Left, header, body, palette)
	}
	if m.helpOpen {
		body := m.output.View()
		return lipgloss.JoinVertical(lipgloss.Left, header, body, m.renderHelpOverlay())
	}

	body := m.output.View()
	prompt := m.renderPrompt()
	return lipgloss.JoinVertical(lipgloss.Left, header, body, prompt, footer)
}

// renderHeader produces the status bar shown at the top of the screen.
func (m Model) renderHeader() string {
	left := fmt.Sprintf("tzcli %s", m.header.Version)
	mid := ""
	if m.header.RepoRoot != "" {
		repoName := shortRepo(m.header.RepoRoot)
		branch := m.header.Branch
		if branch == "" {
			branch = "(no branch)"
		}
		mid = fmt.Sprintf(" repo: %s@%s", repoName, branch)
	}
	tools := ""
	if m.header.ToolsLine != "" {
		tools = " tools: " + m.header.ToolsLine
	}
	drift := ""
	switch {
	case m.header.DriftCount > 0:
		drift = m.chrome.Warn.Render(fmt.Sprintf(" drift: %d", m.header.DriftCount))
	case m.header.DriftCount == 0 && m.header.RepoRoot != "":
		drift = m.chrome.OK.Render(" drift: 0")
	}

	left = m.chrome.Header.Render(left)
	right := m.chrome.HeaderDim.Render(mid + tools + drift)
	bar := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	if w := lipgloss.Width(bar); w < m.width {
		bar += m.chrome.HeaderDim.Render(strings.Repeat(" ", m.width-w))
	}
	return bar
}

// renderPrompt draws the input line with the styled prompt glyph and a
// spinner if a command is in flight.
func (m Model) renderPrompt() string {
	prompt := m.chrome.Prompt.Render("tzcli> ")
	if m.busy {
		spin := m.spin.View()
		busy := m.chrome.HeaderDim.Render(" running: " + m.busyLabel)
		return prompt + spin + busy
	}
	return prompt + m.input.View()
}

// renderFooter draws the keybinding hint line and any inline suggestions.
func (m Model) renderFooter() string {
	hint := m.help.View(m.keys)
	if len(m.suggestions) > 0 {
		hint = m.chrome.Footer.Render("suggestions: "+strings.Join(m.suggestions, " ")) + "\n" + hint
	}
	return m.chrome.Footer.Render(hint)
}

// renderHelpOverlay draws the expanded help block. Press ? again to close.
func (m Model) renderHelpOverlay() string {
	hh := help.New()
	hh.ShowAll = true
	body := hh.View(m.keys)
	return m.chrome.Palette.Render(body)
}

// shortRepo trims a path to its last segment for compactness.
func shortRepo(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}

// welcomeBanner emits a one-time decorative welcome message that shows up
// in the output area when the TUI starts. It now renders the full mnemonic
// menu so the user sees what they can type without having to fish through
// `help` or the palette.
func welcomeBanner(chrome *chromeStyles, version string) tea.Cmd {
	return func() tea.Msg {
		return appendOutputMsg{text: renderStartupMenu(chrome, version)}
	}
}

// appendOutputMsg adds text to the scrollback without invoking a command.
type appendOutputMsg struct{ text string }

// headerRefreshedMsg replaces the header bar in one shot.
type headerRefreshedMsg struct{ info HeaderInfo }

// paletteRefreshedMsg replaces the palette items.
type paletteRefreshedMsg struct{ items []list.Item }

// suggestionsMsg flashes a list of completion candidates in the footer.
type suggestionsMsg struct{ items []string }

// runStartedMsg tells the model a command was kicked off; used to set the
// busy state with a label like the one shown in the prompt.
type runStartedMsg struct{ line string }

// busyClearMsg resets the busy state when the command finishes.
type busyClearMsg struct{}

// quitMsg makes the program exit after history is saved.
type quitMsg struct{}

// keep time imported; we use it in dispatch.go and tests.
var _ = time.Now
