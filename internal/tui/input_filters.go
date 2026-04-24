package tui

// prdNameSeparators are the word-separator runes used by PRD-name editors
// (both FirstTimeSetup's StepPRDName and PRDPicker's new-PRD-name input) for
// Ctrl+Left/Right word jumps. Defined once so the two widgets can't drift.
var prdNameSeparators = []rune{'-', '_'}

// branchNameSeparators are the word-separator runes used by the BranchWarning
// branch-name editor for Ctrl+Left/Right word jumps.
var branchNameSeparators = []rune{'-', '_', '/'}

// filterPRDNameRunes drops any rune outside the allowed PRD-name character
// set ([a-zA-Z0-9_-]). Returns a new slice so the caller can safely forward
// the filtered KeyMsg to the textinput.
func filterPRDNameRunes(runes []rune) []rune {
	filtered := make([]rune, 0, len(runes))
	for _, r := range runes {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterBranchNameRunes drops any rune outside the allowed branch-name
// character set ([a-zA-Z0-9_/-]). Returns a new slice so the caller can safely
// forward the filtered KeyMsg to the textinput.
func filterBranchNameRunes(runes []rune) []rune {
	filtered := make([]rune, 0, len(runes))
	for _, r := range runes {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '/' {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// wordBackward returns the caret position after a word-jump-left from pos,
// treating any rune in seps as a word separator. Mirrors bubbles'
// wordBackward structure (skip separators, then skip non-separators) so
// behavior is predictable next to the built-in key bindings.
func wordBackward(value []rune, pos int, seps []rune) int {
	if pos <= 0 || len(value) == 0 {
		return 0
	}
	i := pos - 1
	for i >= 0 && isSeparator(value[i], seps) {
		i--
	}
	for i >= 0 && !isSeparator(value[i], seps) {
		i--
	}
	return i + 1
}

// wordForward is the forward counterpart of wordBackward.
func wordForward(value []rune, pos int, seps []rune) int {
	n := len(value)
	if pos >= n {
		return n
	}
	i := pos
	for i < n && isSeparator(value[i], seps) {
		i++
	}
	for i < n && !isSeparator(value[i], seps) {
		i++
	}
	return i
}

func isSeparator(r rune, seps []rune) bool {
	for _, s := range seps {
		if r == s {
			return true
		}
	}
	return false
}
