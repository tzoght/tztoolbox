package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Update routes events through the model. Order matters:
//
//  1. Window resize is handled first so children can lay out correctly.
//  2. The command palette eats most keys when open.
//  3. Otherwise key bindings dispatch to the right action.
//  4. Anything that falls through is forwarded to the input/viewport.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	case headerRefreshedMsg:
		m.header = msg.info
		return m, nil

	case paletteRefreshedMsg:
		m.palette.SetItems(msg.items)
		return m, nil

	case appendOutputMsg:
		m.appendBody(msg.text)
		m.output.GotoBottom()
		return m, nil

	case suggestionsMsg:
		m.suggestions = msg.items
		return m, nil

	case runStartedMsg:
		m.busy = true
		m.busyLabel = msg.line
		return m, m.spin.Tick

	case busyClearMsg:
		m.busy = false
		m.busyLabel = ""
		return m, nil

	case commandDoneMsg:
		m.busy = false
		m.busyLabel = ""
		m.appendCommandRecord(msg)
		m.output.GotoBottom()
		return m, m.refreshHeader()

	case tea.KeyMsg:
		// Help overlay closes on '?' or esc.
		if m.helpOpen {
			if key.Matches(msg, m.keys.Help, m.keys.PaletteClose) {
				m.helpOpen = false
				return m, nil
			}
			return m, nil
		}
		// Palette mode: arrow / type / enter goes to the palette.
		if m.paletteOpen {
			switch {
			case key.Matches(msg, m.keys.PaletteClose):
				m.paletteOpen = false
				return m, nil
			case key.Matches(msg, m.keys.Submit):
				if it, ok := m.palette.SelectedItem().(paletteItem); ok {
					m.input.SetValue(it.command + " ")
					m.input.CursorEnd()
				}
				m.paletteOpen = false
				return m, nil
			default:
				var pc tea.Cmd
				m.palette, pc = m.palette.Update(msg)
				return m, pc
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, m.quit()

		case key.Matches(msg, m.keys.Cancel):
			if m.busy {
				// We don't actually cancel running commands in v1; just
				// surface the fact to the user.
				return m, nil
			}
			m.input.SetValue("")
			m.suggestions = nil
			return m, nil

		case key.Matches(msg, m.keys.ClearScreen):
			m.body.Reset()
			m.output.SetContent("")
			return m, nil

		case key.Matches(msg, m.keys.Help):
			m.helpOpen = !m.helpOpen
			return m, nil

		case key.Matches(msg, m.keys.Menu):
			m.appendBody(renderStartupMenu(m.chrome, m.header.Version))
			m.output.GotoBottom()
			return m, nil

		case key.Matches(msg, m.keys.Palette):
			m.paletteOpen = true
			return m, nil

		case key.Matches(msg, m.keys.HistoryUp):
			m.navigateHistory(-1)
			return m, nil

		case key.Matches(msg, m.keys.HistoryDown):
			m.navigateHistory(1)
			return m, nil

		case key.Matches(msg, m.keys.Complete):
			m.applyCompletion()
			return m, nil

		case key.Matches(msg, m.keys.Submit):
			line := strings.TrimSpace(m.input.Value())
			if line == "" {
				return m, nil
			}
			cmds = append(cmds, m.submit(line))
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.ScrollUp):
			m.output.HalfPageUp()
			return m, nil
		case key.Matches(msg, m.keys.ScrollDown):
			m.output.HalfPageDown()
			return m, nil
		case key.Matches(msg, m.keys.ScrollHome):
			m.output.GotoTop()
			return m, nil
		case key.Matches(msg, m.keys.ScrollEnd):
			m.output.GotoBottom()
			return m, nil
		}
	}

	// Forward to spinner/input/viewport so they can update on tick / typing.
	var sc, ic, vc tea.Cmd
	m.spin, sc = m.spin.Update(msg)
	m.input, ic = m.input.Update(msg)
	m.output, vc = m.output.Update(msg)
	cmds = append(cmds, sc, ic, vc)
	return m, tea.Batch(cmds...)
}

// layout sizes the viewport and input to fill the available space below the
// header (1 line) and above the prompt + footer (~3 lines).
func (m *Model) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	headerH := 1
	footerH := 2
	bodyH := m.height - headerH - footerH - 1
	if bodyH < 3 {
		bodyH = 3
	}
	m.output.Width = m.width
	m.output.Height = bodyH
	m.input.Width = m.width - 8
	m.palette.SetSize(m.width-4, m.height-headerH-2)
}

// submit dispatches a single line. Updates history immediately, scrolls to
// bottom, and starts the command goroutine.
func (m *Model) submit(line string) tea.Cmd {
	m.history = pushHistory(m.history, line)
	m.historyIdx = -1
	m.input.SetValue("")
	m.suggestions = nil

	m.appendBody(m.chrome.Prompt.Render("tzcli> ") + m.chrome.PromptCmd.Render(line) + "\n")
	m.output.GotoBottom()

	// Expand REPL-only mnemonics for the local switch below; `runLine`
	// re-applies expansion before talking to Cobra. Doing it here lets
	// single-letter keys like "q" or "c" trigger the local fast-paths.
	expanded := strings.TrimSpace(expandMnemonic(line))
	switch expanded {
	case "quit", "exit", ":q":
		return m.quit()
	case "help":
		return tea.Batch(
			func() tea.Msg { return runStartedMsg{line: line} },
			m.runLine(line),
		)
	case "menu":
		m.appendBody(renderStartupMenu(m.chrome, m.header.Version))
		m.output.GotoBottom()
		return nil
	case "clear":
		m.body.Reset()
		m.output.SetContent("")
		return nil
	}

	return tea.Batch(
		func() tea.Msg { return runStartedMsg{line: line} },
		m.runLine(line),
	)
}

// quit persists history and signals the program loop to exit.
func (m Model) quit() tea.Cmd {
	_ = saveHistory(m.historyPath, m.history)
	return tea.Quit
}

// navigateHistory walks the history stack by delta (-1 = older, +1 = newer).
// Stashes the in-progress draft when starting from the end.
func (m *Model) navigateHistory(delta int) {
	if len(m.history) == 0 {
		return
	}
	if m.historyIdx == -1 {
		m.draft = m.input.Value()
		if delta < 0 {
			m.historyIdx = len(m.history) - 1
			m.input.SetValue(m.history[m.historyIdx])
			m.input.CursorEnd()
		}
		return
	}
	idx := m.historyIdx + delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.history) {
		m.historyIdx = -1
		m.input.SetValue(m.draft)
		m.input.CursorEnd()
		return
	}
	m.historyIdx = idx
	m.input.SetValue(m.history[idx])
	m.input.CursorEnd()
}

// applyCompletion runs Tab completion against the current input.
func (m *Model) applyCompletion() {
	line := m.input.Value()
	root := m.factory()
	if root == nil {
		return
	}
	newLine, suggestions := completeLine(root, line)
	if newLine != line {
		m.input.SetValue(newLine)
		m.input.CursorEnd()
	}
	m.suggestions = suggestions
}

// appendBody concatenates text to the scrollback string and pushes it into
// the viewport. We keep the full string on the model so resize re-wraps
// correctly.
func (m *Model) appendBody(text string) {
	if m.body == nil {
		m.body = &strings.Builder{}
	}
	m.body.WriteString(text)
	m.output.SetContent(m.body.String())
}

// appendCommandRecord prints the captured stdout/stderr (already styled) of
// a finished command, plus a footer line with status + duration.
func (m *Model) appendCommandRecord(msg commandDoneMsg) {
	if msg.output != "" {
		m.appendBody(msg.output)
		if !strings.HasSuffix(msg.output, "\n") {
			m.appendBody("\n")
		}
	}
	d := msg.duration.Round(time.Millisecond)
	footer := ""
	switch {
	case msg.err != nil:
		footer = m.chrome.Err.Render(fmt.Sprintf("error: %v", msg.err))
		if hint := helpHint(msg.head); hint != "" {
			footer += "\n" + m.chrome.Footer.Render(hint)
		}
	default:
		footer = m.chrome.OK.Render("ok")
	}
	footer += "  " + m.chrome.Time.Render(fmt.Sprintf("(%s)", d))
	m.appendBody(footer + "\n\n")
}

// helpHint is rendered under an error footer to nudge the user toward the
// per-subcommand help. We skip it when the head token is itself "help"
// (recursive nudge) or a built-in like "quit" / "clear".
func helpHint(head string) string {
	if head == "" {
		return ""
	}
	switch head {
	case "help", "quit", "exit", "clear", ":q":
		return ""
	}
	return fmt.Sprintf("hint: try '%s --help' or 'help %s' for usage", head, head)
}

// paletteItem is a single row in the command palette.
type paletteItem struct {
	command string
	desc    string
}

func (p paletteItem) Title() string       { return p.command }
func (p paletteItem) Description() string { return p.desc }
func (p paletteItem) FilterValue() string { return p.command + " " + p.desc }

// paletteDelegate gives palette rows minimal styling that matches our chrome.
type paletteDelegate struct {
	chrome *chromeStyles
}

func (d paletteDelegate) Height() int                             { return 2 }
func (d paletteDelegate) Spacing() int                            { return 0 }
func (d paletteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d paletteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(paletteItem)
	if !ok {
		return
	}
	cur := index == m.Index()
	prefix := "  "
	if cur {
		prefix = d.chrome.Prompt.Render("> ")
	}
	title := d.chrome.PromptCmd.Render(it.command)
	desc := d.chrome.Footer.Render(it.desc)
	_, _ = fmt.Fprintf(w, "%s%s\n  %s\n", prefix, title, desc)
}

// builtinPaletteItems is the static minimum the palette ships before the
// factory has been queried. Helpful so Ctrl-K works even on the very first
// keypress before Init's tea.Cmd has fired. We derive it from mnemonics so
// the palette and the startup menu stay in sync.
func builtinPaletteItems() []list.Item {
	items := mnemonicPaletteItems()
	out := make([]list.Item, 0, len(items))
	for _, it := range items {
		out = append(out, it)
	}
	return out
}
