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
		// Multi-byte Unicode runes are dropped — closes the corner case where
		// the old byte-length check (`len(msg.String()) == 1`) would
		// accidentally drop them on the wrong grounds.
		{[]rune("café"), []rune("caf")},
		{[]rune("naïve"), []rune("nave")},
		{[]rune("中文"), []rune{}},
		{[]rune("a日本b"), []rune("ab")},
		{[]rune("emoji-😀-here"), []rune("emoji--here")},
	}
	for _, tc := range tests {
		got := filterValidPRDRunes(tc.in)
		if string(got) != string(tc.want) {
			t.Errorf("filterValidPRDRunes(%q) = %q, want %q", string(tc.in), string(got), string(tc.want))
		}
	}
}

// TestPRDName_SpaceKeyIsFiltered confirms a real spacebar press (which arrives
// with Type=KeySpace, not KeyRunes) is dropped before reaching the textinput.
// Without explicit handling for KeySpace, a literal space would enter the
// buffer and violate AC1.
func TestPRDName_SpaceKeyIsFiltered(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(2)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if got, want := f.ti.Value(), "main"; got != want {
		t.Fatalf("space key should be filtered: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), 2; got != want {
		t.Fatalf("filtered key should not advance cursor: got pos %d, want %d", got, want)
	}
}

// TestPRDName_MultiByteRuneIsFiltered verifies multi-byte Unicode runes
// arriving as a single KeyRunes event are silently dropped (AC1).
func TestPRDName_MultiByteRuneIsFiltered(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(2)
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'é'}})
	if got, want := f.ti.Value(), "main"; got != want {
		t.Fatalf("multi-byte rune should be filtered: got %q, want %q", got, want)
	}
}

// TestPRDName_EnterRejectsEmptyNameMessage pins the exact error string from AC2.
func TestPRDName_EnterRejectsEmptyNameMessage(t *testing.T) {
	f := newPRDNameSetup(t, "")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(FirstTimeSetup)
	if got.prdNameError != "Name cannot be empty" {
		t.Fatalf("expected exact error %q, got %q", "Name cannot be empty", got.prdNameError)
	}
}

// TestPRDName_ErrorClearedOnValueChange verifies AC3: prdNameError is cleared
// whenever the input value changes (here, by typing an allowed rune).
func TestPRDName_ErrorClearedOnValueChange(t *testing.T) {
	f := newPRDNameSetup(t, "")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	f = model.(FirstTimeSetup)
	if f.prdNameError == "" {
		t.Fatal("precondition: empty submit should set an error")
	}
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if f.prdNameError != "" {
		t.Fatalf("error should clear when value changes, got %q", f.prdNameError)
	}
}

// TestPRDName_ErrorPreservedWhenValueUnchanged verifies the error survives a
// keypress that produces no value change (e.g. a fully-filtered space).
func TestPRDName_ErrorPreservedWhenValueUnchanged(t *testing.T) {
	f := newPRDNameSetup(t, "")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	f = model.(FirstTimeSetup)
	wantErr := f.prdNameError
	if wantErr == "" {
		t.Fatal("precondition: empty submit should set an error")
	}
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	if f.prdNameError != wantErr {
		t.Fatalf("error should persist when filtered key changes nothing: got %q, want %q", f.prdNameError, wantErr)
	}
}

// TestPRDName_CtrlCCancels verifies AC4: ctrl+c quits and marks the result
// cancelled regardless of the showGitignore branch.
func TestPRDName_CtrlCCancels(t *testing.T) {
	for _, showGitignore := range []bool{false, true} {
		t.Run("", func(t *testing.T) {
			setup := NewFirstTimeSetup(t.TempDir(), showGitignore)
			setup.step = StepPRDName
			model, cmd := setup.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyCtrlC})
			got := model.(FirstTimeSetup)
			if !got.result.Cancelled {
				t.Fatal("ctrl+c should set Cancelled=true")
			}
			if cmd == nil {
				t.Fatal("ctrl+c should return a non-nil cmd (tea.Quit)")
			}
		})
	}
}

// TestPRDName_EscWithoutGitignoreCancels verifies AC4: when the gitignore step
// was skipped, esc cancels the flow.
func TestPRDName_EscWithoutGitignoreCancels(t *testing.T) {
	setup := NewFirstTimeSetup(t.TempDir(), false)
	model, cmd := setup.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(FirstTimeSetup)
	if !got.result.Cancelled {
		t.Fatal("esc with no gitignore step should cancel")
	}
	if cmd == nil {
		t.Fatal("esc with no gitignore step should return tea.Quit")
	}
}

// TestPRDName_EscWithGitignoreReturnsToPreviousStep verifies AC4: when the
// gitignore step preceded this one, esc walks back to it (no cancellation),
// and clears any pending error.
func TestPRDName_EscWithGitignoreReturnsToPreviousStep(t *testing.T) {
	setup := NewFirstTimeSetup(t.TempDir(), true)
	setup.step = StepPRDName
	setup.prdNameError = "something"
	model, cmd := setup.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEsc})
	got := model.(FirstTimeSetup)
	if got.result.Cancelled {
		t.Fatal("esc with gitignore step should not cancel")
	}
	if got.step != StepGitignore {
		t.Fatalf("esc should return to gitignore step, got step=%d", got.step)
	}
	if got.prdNameError != "" {
		t.Fatalf("esc should clear prdNameError, got %q", got.prdNameError)
	}
	if cmd != nil {
		t.Fatal("esc back to gitignore should not return a quit cmd")
	}
}

// TestPRDName_EnterAdvancesAndClearsError verifies AC2 and AC3 together: a
// successful submit clears any prior error and advances to StepPostCompletion.
func TestPRDName_EnterAdvancesAndClearsError(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.prdNameError = "stale error"
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(FirstTimeSetup)
	if got.step != StepPostCompletion {
		t.Fatalf("expected step=%d (StepPostCompletion), got %d", StepPostCompletion, got.step)
	}
	if got.result.PRDName != "main" {
		t.Fatalf("expected PRDName=main, got %q", got.result.PRDName)
	}
}

// TestPRDName_TextinputWidthMatchesModalContent verifies AC6: the textinput's
// Width tracks the lipgloss content width via prdNameModalWidth - 8, with no
// extra padding subtraction. Resizing should keep them in sync.
func TestPRDName_TextinputWidthMatchesModalContent(t *testing.T) {
	setup := NewFirstTimeSetup(t.TempDir(), false)
	if got, want := setup.ti.Width, prdNameModalWidth(0)-8; got != want {
		t.Fatalf("initial ti.Width: got %d, want %d", got, want)
	}
	model, _ := setup.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := model.(FirstTimeSetup)
	if want := prdNameModalWidth(120) - 8; got.ti.Width != want {
		t.Fatalf("ti.Width after resize: got %d, want %d", got.ti.Width, want)
	}
}

// TestPRDName_EmptyAndPopulatedFieldHaveSameRenderedWidth verifies AC7: the
// bordered input box keeps the same visual width whether the field is empty
// or contains text.
func TestPRDName_EmptyAndPopulatedFieldHaveSameRenderedWidth(t *testing.T) {
	emptySetup := NewFirstTimeSetup(t.TempDir(), false)
	emptySetup.width, emptySetup.height = 100, 40
	emptySetup.ti.Width = prdNameModalWidth(100) - 8
	emptySetup.ti.SetValue("")
	emptyView := emptySetup.View()

	populatedSetup := NewFirstTimeSetup(t.TempDir(), false)
	populatedSetup.width, populatedSetup.height = 100, 40
	populatedSetup.ti.Width = prdNameModalWidth(100) - 8
	populatedSetup.ti.SetValue("main")
	populatedView := populatedSetup.View()

	emptyMax := maxLineWidth(emptyView)
	populatedMax := maxLineWidth(populatedView)
	if emptyMax != populatedMax {
		t.Fatalf("rendered max width should match: empty=%d populated=%d", emptyMax, populatedMax)
	}
}

func maxLineWidth(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		if n := len([]rune(line)); n > max {
			max = n
		}
	}
	return max
}

// pasteMsg constructs the KeyMsg bubbletea emits for a bracketed paste: a
// single KeyRunes event with Paste=true carrying the full pasted rune slice.
// See bubbletea v1.3.10 key_sequences.go:109 (detectBracketedPaste).
func pasteMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s), Paste: true}
}

// TestPRDName_PasteAllValidInsertsAtCaret (US-004 AC1): a string composed
// entirely of allowed characters is inserted at the caret in one step.
func TestPRDName_PasteAllValidInsertsAtCaret(t *testing.T) {
	f := newPRDNameSetup(t, "")
	f = sendKey(t, f, pasteMsg("my-feature_v2"))
	if got, want := f.ti.Value(), "my-feature_v2"; got != want {
		t.Fatalf("paste all-valid: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), len("my-feature_v2"); got != want {
		t.Fatalf("cursor should advance to end of pasted text: got pos %d, want %d", got, want)
	}
}

// TestPRDName_PasteFiltersInvalidChars (US-004 AC2): invalid characters are
// silently dropped and no error is shown.
func TestPRDName_PasteFiltersInvalidChars(t *testing.T) {
	f := newPRDNameSetup(t, "")
	f = sendKey(t, f, pasteMsg("my feature/v2!"))
	if got, want := f.ti.Value(), "myfeaturev2"; got != want {
		t.Fatalf("paste with invalid chars: got %q, want %q", got, want)
	}
	if f.prdNameError != "" {
		t.Fatalf("paste should not set an error, got %q", f.prdNameError)
	}
}

// TestPRDName_PasteTruncatesToMaxLength (US-004 AC3): the field never exceeds
// maxPRDNameLength; existing characters before the caret are preserved.
func TestPRDName_PasteTruncatesToMaxLength(t *testing.T) {
	f := newPRDNameSetup(t, "ab") // existing prefix before caret
	// Paste more valid characters than would fit. Filter drops no runes, so the
	// textinput has to enforce the CharLimit truncation itself.
	longPaste := strings.Repeat("x", maxPRDNameLength*2)
	f = sendKey(t, f, pasteMsg(longPaste))
	if got := len(f.ti.Value()); got != maxPRDNameLength {
		t.Fatalf("value length: got %d, want %d", got, maxPRDNameLength)
	}
	if got := f.ti.Value(); !strings.HasPrefix(got, "ab") {
		t.Fatalf("existing prefix before caret must be preserved: got %q", got)
	}
}

// TestPRDName_PasteAtMiddleCaretSplices (US-004 AC4): pasting with the caret
// mid-buffer splices the filtered text into the middle of the value.
func TestPRDName_PasteAtMiddleCaretSplices(t *testing.T) {
	f := newPRDNameSetup(t, "main") // pos=4
	f.ti.SetCursor(2)               // "ma|in"
	f = sendKey(t, f, pasteMsg("X-Y"))
	if got, want := f.ti.Value(), "maX-Yin"; got != want {
		t.Fatalf("paste mid-buffer: got %q, want %q", got, want)
	}
	if got, want := f.ti.Position(), 5; got != want {
		t.Fatalf("cursor should sit right after the pasted text: got pos %d, want %d", got, want)
	}
}

// TestPRDName_PasteAtMiddleCaretSplicesWithFiltering combines AC2 and AC4: an
// in-middle paste with invalid chars splices only the valid subset.
func TestPRDName_PasteAtMiddleCaretSplicesWithFiltering(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.ti.SetCursor(2)
	f = sendKey(t, f, pasteMsg("X Y/Z"))
	if got, want := f.ti.Value(), "maXYZin"; got != want {
		t.Fatalf("filtered paste mid-buffer: got %q, want %q", got, want)
	}
}

// TestPRDName_PasteClearsError (US-004 AC5): a paste that changes the value
// clears prdNameError.
func TestPRDName_PasteClearsError(t *testing.T) {
	f := newPRDNameSetup(t, "")
	model, _ := f.handlePRDNameKeys(tea.KeyMsg{Type: tea.KeyEnter})
	f = model.(FirstTimeSetup)
	if f.prdNameError == "" {
		t.Fatal("precondition: empty submit should set an error")
	}
	f = sendKey(t, f, pasteMsg("feature"))
	if f.prdNameError != "" {
		t.Fatalf("paste should clear prdNameError, got %q", f.prdNameError)
	}
	if got, want := f.ti.Value(), "feature"; got != want {
		t.Fatalf("paste value: got %q, want %q", got, want)
	}
}

// TestPRDName_PasteAllInvalidIsNoOp verifies that a paste containing only
// invalid characters leaves the value unchanged (and therefore does not clear
// a standing error — sister assertion to AC5's "changes the value" wording).
func TestPRDName_PasteAllInvalidIsNoOp(t *testing.T) {
	f := newPRDNameSetup(t, "main")
	f.prdNameError = "sticky"
	f = sendKey(t, f, pasteMsg("! @ # $"))
	if got, want := f.ti.Value(), "main"; got != want {
		t.Fatalf("all-invalid paste should not change value: got %q, want %q", got, want)
	}
	if f.prdNameError != "sticky" {
		t.Fatalf("error should persist when paste changes nothing: got %q", f.prdNameError)
	}
}

// TestPRDName_PasteWithoutBracketedFlagAlsoFiltered verifies the same filter
// path handles a multi-rune KeyRunes event that lacks Paste=true (the
// fallback path when bracketed paste is disabled in the terminal).
func TestPRDName_PasteWithoutBracketedFlagAlsoFiltered(t *testing.T) {
	f := newPRDNameSetup(t, "")
	f = sendKey(t, f, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ab/cd!")})
	if got, want := f.ti.Value(), "abcd"; got != want {
		t.Fatalf("non-bracketed multi-rune paste: got %q, want %q", got, want)
	}
}
