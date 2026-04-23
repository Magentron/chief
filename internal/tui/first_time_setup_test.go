package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

// newPRDNameSetup returns a FirstTimeSetup positioned on the PRD-name step
// with the textinput pre-populated to value with the cursor at end.
func newPRDNameSetup(t *testing.T, value string) FirstTimeSetup {
	t.Helper()
	setup := NewFirstTimeSetup(t.TempDir(), false)
	setup.ti.SetValue(value)
	setup.ti.CursorEnd()
	return *setup
}

func sendKey(t *testing.T, f FirstTimeSetup, msg tea.KeyMsg) FirstTimeSetup {
	t.Helper()
	model, _ := f.handlePRDNameKeys(msg)
	got, ok := model.(FirstTimeSetup)
	if !ok {
		t.Fatalf("expected FirstTimeSetup model, got %T", model)
	}
	return got
}

func TestPRDName_InitialCursorAtEnd(t *testing.T) {
	setup := NewFirstTimeSetup(t.TempDir(), false)
	if got, want := setup.ti.Value(), "main"; got != want {
		t.Fatalf("initial value: got %q, want %q", got, want)
	}
	if got, want := setup.ti.Position(), len("main"); got != want {
		t.Fatalf("initial cursor position: got %d, want %d", got, want)
	}
}

func TestPRDName_LeftArrowMovesCaretLeft(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyLeft})
	if got, want := f.ti.Position(), 3; got != want {
		t.Fatalf("after left: got pos %d, want %d", got, want)
	}
	if got, want := f.ti.Value(), "main"; got != want {
		t.Fatalf("value should be unchanged: got %q, want %q", got, want)
	}
}

func TestPRDName_LeftArrowAtPositionZeroIsNoOp(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(0)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyLeft})
	if got, want := f.ti.Position(), 0; got != want {
		t.Fatalf("left at pos 0 should be no-op: got pos %d, want %d", got, want)
	}
}

func TestPRDName_RightArrowMovesCaretRight(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(0)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRight})
	if got, want := f.ti.Position(), 1; got != want {
		t.Fatalf("after right: got pos %d, want %d", got, want)
	}
}

func TestPRDName_RightArrowAtEndIsNoOp(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4 (end)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRight})
	if got, want := f.ti.Position(), 4; got != want {
		t.Fatalf("right at end should be no-op: got pos %d, want %d", got, want)
	}
}

func TestPRDName_HomeJumpsToStart(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyHome})
	if got, want := f.ti.Position(), 0; got != want {
		t.Fatalf("after home: got pos %d, want %d", got, want)
	}
}

func TestPRDName_EndJumpsToEnd(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(0)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyEnd})
	if got, want := f.ti.Position(), 4; got != want {
		t.Fatalf("after end: got pos %d, want %d", got, want)
	}
}

func TestPRDName_CtrlLeftJumpsWordLeft(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4, no whitespace → one word
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyCtrlLeft})
	if got, want := f.ti.Position(), 0; got != want {
		t.Fatalf("after ctrl+left: got pos %d, want %d", got, want)
	}
}

func TestPRDName_CtrlRightJumpsWordRight(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(0)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyCtrlRight})
	if got, want := f.ti.Position(), 4; got != want {
		t.Fatalf("after ctrl+right: got pos %d, want %d", got, want)
	}
}

func TestPRDName_TypeInsertsAtCaret(t *testing.T) {
	f := newPRDNameSetup(t, "main") // value=main, pos=4
	f.ti.SetCursor(2)               // between 'a' and 'i' → "ma|in"
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if got, want := f.ti.Value(), "maXin"; got != want {
		t.Fatalf("after insert at caret: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), 3; got != want {
		t.Fatalf("cursor should advance past inserted rune: got pos %d, want %d", got, want)
	}
}

func TestPRDName_TypeDisallowedRuneIsFiltered(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(2)
	// Mix of allowed ('Y') and disallowed (' ', '!').
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y', ' ', '!'}})
	if got, want := f.ti.Value(), "maYin"; got != want {
		t.Fatalf("only allowed runes should be inserted: got %q, want %q", got, want)
	}
}

func TestPRDName_BackspaceDeletesCharBeforeCaret(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4
	f.ti.SetCursor(2)               // "ma|in" → backspace deletes 'a'
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyBackspace})
	if got, want := f.ti.Value(), "min"; got != want {
		t.Fatalf("backspace at caret: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), 1; got != want {
		t.Fatalf("cursor should move left after backspace: got pos %d, want %d", got, want)
	}
}

func TestPRDName_BackspaceAtPositionZeroIsNoOp(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(0)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyBackspace})
	if got, want := f.ti.Value(), "main"; got != want {
		t.Fatalf("backspace at pos 0 should be no-op: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), 0; got != want {
		t.Fatalf("cursor at 0 should stay at 0: got pos %d, want %d", got, want)
	}
}

func TestPRDName_ViewRendersVisibleCaret(t *testing.T) {
	// The visible caret comes from bubbles' cursor.Model rendering a styled
	// block over the character at the cursor position. We can't reliably assert
	// on ANSI escapes in tests (lipgloss strips styling when stdout isn't a
	// TTY), so we verify the preconditions that make the caret visible at
	// runtime: the input is focused, and the cursor is in blink mode (which
	// renders a reverse-video block when focused).
	f := newPRDNameSetup(t, "main")
	if !f.ti.Focused() {
		t.Fatal("textinput must be focused for the caret to render")
	}
	if f.ti.Cursor.Mode() != cursor.CursorBlink {
		t.Fatalf("cursor mode must be CursorBlink for a visible caret, got %v", f.ti.Cursor.Mode())
	}
	// View() must contain the input value, confirming the field is rendered.
	if !strings.Contains(f.ti.View(), "main") {
		t.Fatalf("View() should render the input value, got %q", f.ti.View())
	}
}

func TestPRDName_EnterClearsErrorAndAdvances(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(FirstTimeSetup)
	if got.step != StepPostCompletion {
		t.Fatalf("enter should advance to post-completion step, got %d", got.step)
	}
	if got.result.PRDName != "main" {
		t.Fatalf("expected result.PRDName=main, got %q", got.result.PRDName)
	}
}

func TestPRDName_EnterRejectsEmptyName(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetValue("")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(FirstTimeSetup)
	if got.step != StepPRDName {
		t.Fatalf("empty name should not advance: step=%d", got.step)
	}
	if got.prdNameError == "" {
		t.Fatal("expected an error message for empty name")
	}
}

func TestFilterValidPRDRunes(t *testing.T) {
	tests := []struct {
		in   []rune
		want []rune
	}{
		{[]rune("abcXYZ"), []rune("abcXYZ")},
		{[]rune("a-b_c"), []rune("a-b_c")},
		{[]rune("01234"), []rune("01234")},
		{[]rune("a b!c"), []rune("abc")},
		{[]rune(""), []rune{}},
	}
	for _, tc := range tests {
		got := filterValidPRDRunes(tc.in)
		if string(got) != string(tc.want) {
			t.Errorf("filterValidPRDRunes(%q) = %q, want %q", string(tc.in), string(got), string(tc.want))
		}
	}
}
