package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyPress is a tiny helper for crafting tea.KeyMsg values for tests.
func keyPress(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func runesPress(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestUpdateWindowSizeLaysOutChildren(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	got := updated.(Model)
	if got.width != 100 || got.height != 30 {
		t.Fatalf("expected width/height to be set, got %dx%d", got.width, got.height)
	}
	if got.output.Width != 100 {
		t.Fatalf("expected viewport.Width=100, got %d", got.output.Width)
	}
}

func TestCtrlLClearsBody(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	m.appendBody("some prior output\n")
	if m.body.Len() == 0 {
		t.Fatal("expected body to be non-empty")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	got := updated.(Model)
	if got.body.Len() != 0 {
		t.Fatalf("expected ctrl+l to clear body, still: %q", got.body.String())
	}
}

func TestCtrlDQuits(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Fatal("expected non-nil quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestF1ReprintsMenu(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{}), Version: "v0.0.0-test"})
	m, _ = applyResize(m, 100, 30)
	// Ensure the body starts empty so we know F1 added content.
	m.body.Reset()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF1})
	body := updated.(Model).body.String()
	if !strings.Contains(body, "welcome to tzcli v0.0.0-test") {
		t.Fatalf("F1 should reprint the menu, body=%q", body)
	}
	if !strings.Contains(body, "[s]") || !strings.Contains(body, "sync") {
		t.Fatalf("expected mnemonic rows in menu output, got %q", body)
	}
}

func TestCtrlGReprintsMenu(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{}), Version: "v0.0.0-test"})
	m, _ = applyResize(m, 100, 30)
	m.body.Reset()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	body := updated.(Model).body.String()
	if !strings.Contains(body, "welcome to tzcli") {
		t.Fatalf("ctrl+g should reprint the menu, body=%q", body)
	}
}

func TestMenuMnemonicReprintsMenu(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{}), Version: "v0.0.0-test"})
	m, _ = applyResize(m, 100, 30)
	m.body.Reset()
	m.input.SetValue("m")

	updated, _ := m.Update(keyPress(tea.KeyEnter))
	body := updated.(Model).body.String()
	if !strings.Contains(body, "welcome to tzcli") {
		t.Fatalf("typing 'm'+Enter should reprint the menu, body=%q", body)
	}
	if updated.(Model).input.Value() != "" {
		t.Fatal("input should be cleared after running 'm'")
	}
}

func TestPaletteOpensOnCtrlK(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	if !updated.(Model).paletteOpen {
		t.Fatal("expected palette to be open")
	}
}

func TestEnterEmptyInputIsNoop(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	updated, cmd := m.Update(keyPress(tea.KeyEnter))
	if cmd != nil {
		t.Fatalf("expected nil cmd for empty enter, got %T", cmd())
	}
	if updated.(Model).busy {
		t.Fatal("model should not be busy after empty enter")
	}
}

func TestCommandDoneAppendsRecord(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	updated, _ := m.Update(commandDoneMsg{
		line:   "echo hi",
		output: "hi\n",
	})
	got := updated.(Model)
	body := got.body.String()
	if !strings.Contains(body, "hi") {
		t.Fatalf("expected output to be appended, got %q", body)
	}
	if !strings.Contains(body, "ok") && !strings.Contains(body, "error:") {
		t.Fatalf("expected status footer in body, got %q", body)
	}
}

func TestRunStartedSetsBusy(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	updated, _ := m.Update(runStartedMsg{line: "doctor"})
	got := updated.(Model)
	if !got.busy || got.busyLabel != "doctor" {
		t.Fatalf("expected busy=true, label='doctor'; got %+v", got)
	}
}

func TestHistoryNavigationCyclesEntries(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	m.history = []string{"sync", "doctor", "installed"}
	// Press up: should land on "installed"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := updated.(Model).input.Value(); got != "installed" {
		t.Fatalf("expected 'installed', got %q", got)
	}
}

func TestSubmitNonEmptyTriggersBatch(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	m.input.SetValue("echo hi")
	updated, cmd := m.Update(keyPress(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("expected non-nil cmd from submit")
	}
	if updated.(Model).input.Value() != "" {
		t.Fatal("expected input to be cleared after submit")
	}
	if len(updated.(Model).history) == 0 {
		t.Fatal("expected history to record the submission")
	}
}

// TestUpdatePreservesBuilderAcrossCopies guards against the
// "non-zero Builder copied by value" panic. Bubbletea's Update returns the
// model by value, so the body must be a pointer to survive multiple Update
// roundtrips after writes.
func TestUpdatePreservesBuilderAcrossCopies(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Update panicked across copies: %v", r)
		}
	}()

	for i := 0; i < 5; i++ {
		updated, _ := m.Update(commandDoneMsg{
			line:   "echo hi",
			output: "hi\n",
		})
		m = updated.(Model)
		// Now feed another Update; this is the path the runtime panic took.
		updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		m = updated.(Model)
	}
	if m.body.Len() == 0 {
		t.Fatal("expected body to retain content after multiple Update copies")
	}
}

func TestRunesTypeIntoInput(t *testing.T) {
	m := NewModel(Options{Factory: makeFactory(&strings.Builder{})})
	m, _ = applyResize(m, 80, 24)
	updated, _ := m.Update(runesPress("hello"))
	if got := updated.(Model).input.Value(); got != "hello" {
		t.Fatalf("expected 'hello' in input, got %q", got)
	}
}

// applyResize is a small helper used across tests so each path doesn't have
// to remember to size the model before running other key events.
func applyResize(m Model, w, h int) (Model, tea.Cmd) {
	updated, c := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return updated.(Model), c
}
