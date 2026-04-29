package tui

import (
	"strings"
	"testing"

	"github.com/Magentron/chief/internal/config"
)

// Item layout after LoadFromConfig — kept in sync with settings.go and the
// YAML grouping in docs/reference/configuration.md:
//
//	0: agent.watchdogTimeout       (string)
//	1: worktree.setup              (string)
//	2: worktree.alwaysPrompt       (bool)
//	3: worktree.promptBranchPattern (string, regex-validated)
//	4: bash.timeout                (string, duration-validated)
//	5: onComplete.push             (bool)
//	6: onComplete.createPR         (bool)

func TestSettingsOverlay_LoadFromConfig(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Agent: config.AgentConfig{WatchdogTimeout: "20m"},
		Worktree: config.WorktreeConfig{
			Setup:               "npm install",
			AlwaysPrompt:        false,
			PromptBranchPattern: "^(main|master)$",
		},
		Bash: config.BashConfig{Timeout: "30s"},
		OnComplete: config.OnCompleteConfig{
			Push:     true,
			CreatePR: false,
		},
	}
	s.LoadFromConfig(cfg)

	if len(s.items) != 7 {
		t.Fatalf("expected 7 items, got %d", len(s.items))
	}
	if s.items[0].Key != "agent.watchdogTimeout" || s.items[0].StringVal != "20m" {
		t.Errorf("agent.watchdogTimeout item: got key=%s val=%s", s.items[0].Key, s.items[0].StringVal)
	}
	if s.items[1].Key != "worktree.setup" || s.items[1].StringVal != "npm install" {
		t.Errorf("worktree.setup item: got key=%s val=%s", s.items[1].Key, s.items[1].StringVal)
	}
	if s.items[2].Key != "worktree.alwaysPrompt" || s.items[2].BoolVal {
		t.Errorf("worktree.alwaysPrompt item: got key=%s val=%v", s.items[2].Key, s.items[2].BoolVal)
	}
	if s.items[3].Key != "worktree.promptBranchPattern" || s.items[3].StringVal != "^(main|master)$" {
		t.Errorf("worktree.promptBranchPattern item: got key=%s val=%s", s.items[3].Key, s.items[3].StringVal)
	}
	if s.items[4].Key != "bash.timeout" || s.items[4].StringVal != "30s" {
		t.Errorf("bash.timeout item: got key=%s val=%s", s.items[4].Key, s.items[4].StringVal)
	}
	if s.items[5].Key != "onComplete.push" || !s.items[5].BoolVal {
		t.Errorf("onComplete.push item: got key=%s val=%v", s.items[5].Key, s.items[5].BoolVal)
	}
	if s.items[6].Key != "onComplete.createPR" || s.items[6].BoolVal {
		t.Errorf("onComplete.createPR item: got key=%s val=%v", s.items[6].Key, s.items[6].BoolVal)
	}
	if s.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0, got %d", s.selectedIndex)
	}
}

func TestSettingsOverlay_ApplyToConfig(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := config.Default()
	s.LoadFromConfig(cfg)

	s.items[0].StringVal = "20m"             // agent.watchdogTimeout
	s.items[1].StringVal = "go mod download" // worktree.setup
	s.items[2].BoolVal = true                // worktree.alwaysPrompt
	s.items[3].StringVal = "^release/.*$"    // worktree.promptBranchPattern
	s.items[4].StringVal = "30s"             // bash.timeout
	s.items[5].BoolVal = true                // onComplete.push
	s.items[6].BoolVal = true                // onComplete.createPR

	resultCfg := config.Default()
	s.ApplyToConfig(resultCfg)

	if resultCfg.Agent.WatchdogTimeout != "20m" {
		t.Errorf("expected agent.watchdogTimeout='20m', got '%s'", resultCfg.Agent.WatchdogTimeout)
	}
	if resultCfg.Worktree.Setup != "go mod download" {
		t.Errorf("expected setup='go mod download', got '%s'", resultCfg.Worktree.Setup)
	}
	if !resultCfg.Worktree.AlwaysPrompt {
		t.Error("expected alwaysPrompt=true")
	}
	if resultCfg.Worktree.PromptBranchPattern != "^release/.*$" {
		t.Errorf("expected promptBranchPattern='^release/.*$', got '%s'", resultCfg.Worktree.PromptBranchPattern)
	}
	if resultCfg.Bash.Timeout != "30s" {
		t.Errorf("expected bash.timeout='30s', got '%s'", resultCfg.Bash.Timeout)
	}
	if !resultCfg.OnComplete.Push {
		t.Error("expected push=true")
	}
	if !resultCfg.OnComplete.CreatePR {
		t.Error("expected createPR=true")
	}
}

func TestSettingsOverlay_Navigation(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	if s.selectedIndex != 0 {
		t.Fatalf("expected initial index=0, got %d", s.selectedIndex)
	}

	for i := 1; i <= 6; i++ {
		s.MoveDown()
		if s.selectedIndex != i {
			t.Errorf("expected index=%d after MoveDown, got %d", i, s.selectedIndex)
		}
	}

	// Can't go beyond last item
	s.MoveDown()
	if s.selectedIndex != 6 {
		t.Errorf("expected index=6 (clamped), got %d", s.selectedIndex)
	}

	s.MoveUp()
	if s.selectedIndex != 5 {
		t.Errorf("expected index=5 after MoveUp, got %d", s.selectedIndex)
	}

	for i := 0; i < 10; i++ {
		s.MoveUp()
	}
	if s.selectedIndex != 0 {
		t.Errorf("expected index=0 (clamped), got %d", s.selectedIndex)
	}
}

// moveTo positions the cursor on the item at the given index.
func moveTo(s *SettingsOverlay, idx int) {
	s.selectedIndex = 0
	for i := 0; i < idx; i++ {
		s.MoveDown()
	}
}

func TestSettingsOverlay_ToggleBool(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		OnComplete: config.OnCompleteConfig{Push: false},
	}
	s.LoadFromConfig(cfg)
	moveTo(s, 5) // onComplete.push

	key, val := s.ToggleBool()
	if key != "onComplete.push" {
		t.Errorf("expected key='onComplete.push', got '%s'", key)
	}
	if !val {
		t.Error("expected val=true after toggle")
	}

	_, val = s.ToggleBool()
	if val {
		t.Error("expected val=false after second toggle")
	}
}

func TestSettingsOverlay_ToggleBool_OnStringItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	// Selected item is "Watchdog timeout" (string type, index 0).
	key, _ := s.ToggleBool()
	if key != "" {
		t.Errorf("expected empty key for string item toggle, got '%s'", key)
	}
}

func TestSettingsOverlay_RevertToggle(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		OnComplete: config.OnCompleteConfig{Push: false},
	}
	s.LoadFromConfig(cfg)
	moveTo(s, 5) // onComplete.push
	s.ToggleBool()
	if !s.items[5].BoolVal {
		t.Fatal("expected true after toggle")
	}

	s.RevertToggle()
	if s.items[5].BoolVal {
		t.Error("expected false after revert")
	}
}

func TestSettingsOverlay_BashTimeoutValidation(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 4) // bash.timeout
	if s.GetSelectedItem().Key != "bash.timeout" {
		t.Fatalf("setup error: expected bash.timeout selected, got %q", s.GetSelectedItem().Key)
	}

	// Invalid duration: edit should be rejected, edit mode preserved.
	s.StartEditing()
	for _, ch := range "5minutes" {
		s.AddEditChar(ch)
	}
	if err := s.ConfirmEdit(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
	if !s.IsEditing() {
		t.Fatal("expected to remain in edit mode for invalid duration")
	}
	if s.editError == "" {
		t.Error("expected editError to be set for invalid duration")
	}
	if s.GetSelectedItem().StringVal != "" {
		t.Errorf("expected stored value to remain unchanged, got %q", s.GetSelectedItem().StringVal)
	}

	// Correct the buffer to a valid value: edit accepted, error cleared,
	// surrounding whitespace trimmed.
	s.editBuffer = "  30s  "
	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("ConfirmEdit unexpected error: %v", err)
	}
	if s.IsEditing() {
		t.Error("expected to exit edit mode after valid duration")
	}
	if s.editError != "" {
		t.Errorf("expected editError cleared, got %q", s.editError)
	}
	if s.GetSelectedItem().StringVal != "30s" {
		t.Errorf("expected stored value '30s', got %q", s.GetSelectedItem().StringVal)
	}
}

func TestSettingsOverlay_BashTimeoutEmptyAccepted(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(&config.Config{Bash: config.BashConfig{Timeout: "5m"}})
	moveTo(s, 4) // bash.timeout
	if s.GetSelectedItem().Key != "bash.timeout" {
		t.Fatalf("setup error: expected bash.timeout selected, got %q", s.GetSelectedItem().Key)
	}

	s.StartEditing()
	s.editBuffer = ""
	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("expected empty value to be accepted, got error: %v", err)
	}
	if s.IsEditing() {
		t.Error("expected empty value to be accepted")
	}
	if s.GetSelectedItem().StringVal != "" {
		t.Errorf("expected stored value '', got %q", s.GetSelectedItem().StringVal)
	}
}

func TestSettingsOverlay_BashTimeoutNegativeRejected(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 4) // bash.timeout
	s.StartEditing()
	for _, ch := range "-10s" {
		s.AddEditChar(ch)
	}
	if err := s.ConfirmEdit(); err == nil {
		t.Fatal("expected negative duration to be rejected")
	}
	if !s.IsEditing() {
		t.Error("expected negative duration to be rejected")
	}
	if s.editError == "" {
		t.Error("expected editError set for negative duration")
	}
}

func TestSettingsOverlay_AgentWatchdogTimeoutValidation(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	if s.GetSelectedItem().Key != "agent.watchdogTimeout" {
		t.Fatalf("setup error: expected agent.watchdogTimeout selected, got %q", s.GetSelectedItem().Key)
	}

	s.StartEditing()
	for _, ch := range "10minutes" {
		s.AddEditChar(ch)
	}
	if err := s.ConfirmEdit(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
	if !s.IsEditing() {
		t.Fatal("expected to remain in edit mode for invalid duration")
	}
	if s.editError == "" {
		t.Error("expected editError to be set for invalid duration")
	}
	if s.GetSelectedItem().StringVal != "" {
		t.Errorf("expected stored value to remain unchanged, got %q", s.GetSelectedItem().StringVal)
	}

	s.editBuffer = "  20m  "
	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("ConfirmEdit unexpected error: %v", err)
	}
	if s.IsEditing() {
		t.Error("expected valid duration to be accepted")
	}
	if s.GetSelectedItem().StringVal != "20m" {
		t.Errorf("expected stored value '20m', got %q", s.GetSelectedItem().StringVal)
	}
}

func TestSettingsOverlay_AgentWatchdogTimeoutNegativeRejected(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	if s.GetSelectedItem().Key != "agent.watchdogTimeout" {
		t.Fatalf("setup error: expected agent.watchdogTimeout selected, got %q", s.GetSelectedItem().Key)
	}
	s.StartEditing()
	for _, ch := range "-5m" {
		s.AddEditChar(ch)
	}
	if err := s.ConfirmEdit(); err == nil {
		t.Fatal("expected negative duration to be rejected")
	}
	if !s.IsEditing() {
		t.Error("expected negative duration to be rejected")
	}
	if s.editError == "" {
		t.Error("expected editError set for negative duration")
	}
}

func TestSettingsOverlay_StringEditing(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 1) // worktree.setup — accepts arbitrary strings
	if s.IsEditing() {
		t.Fatal("should not be editing initially")
	}

	s.StartEditing()
	if !s.IsEditing() {
		t.Fatal("should be editing after StartEditing")
	}
	if s.editBuffer != "" {
		t.Errorf("expected empty edit buffer, got '%s'", s.editBuffer)
	}

	s.AddEditChar('n')
	s.AddEditChar('p')
	s.AddEditChar('m')
	if s.editBuffer != "npm" {
		t.Errorf("expected 'npm', got '%s'", s.editBuffer)
	}

	s.DeleteEditChar()
	if s.editBuffer != "np" {
		t.Errorf("expected 'np' after delete, got '%s'", s.editBuffer)
	}

	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("ConfirmEdit unexpected error: %v", err)
	}
	if s.IsEditing() {
		t.Fatal("should not be editing after ConfirmEdit")
	}
	if s.items[1].StringVal != "np" {
		t.Errorf("expected StringVal='np', got '%s'", s.items[1].StringVal)
	}
}

func TestSettingsOverlay_CancelEdit(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Worktree: config.WorktreeConfig{Setup: "original"},
	}
	s.LoadFromConfig(cfg)
	moveTo(s, 1) // worktree.setup

	s.StartEditing()
	s.AddEditChar('x')
	s.CancelEdit()

	if s.IsEditing() {
		t.Fatal("should not be editing after CancelEdit")
	}
	if s.items[1].StringVal != "original" {
		t.Errorf("expected 'original' preserved, got '%s'", s.items[1].StringVal)
	}
}

func TestSettingsOverlay_StartEditingOnBoolItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 2) // worktree.alwaysPrompt (bool)

	s.StartEditing()
	if s.IsEditing() {
		t.Error("should not start editing on a bool item")
	}
}

func TestSettingsOverlay_ConfirmEdit_InvalidRegex(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 3) // worktree.promptBranchPattern

	original := s.items[3].StringVal
	s.StartEditing()
	for s.editBuffer != "" {
		s.DeleteEditChar()
	}
	for _, ch := range "[bad" {
		s.AddEditChar(ch)
	}

	err := s.ConfirmEdit()
	if err == nil {
		t.Fatal("expected error from ConfirmEdit on invalid regex, got nil")
	}
	if !s.IsEditing() {
		t.Error("expected to remain in edit mode after rejection")
	}
	if s.items[3].StringVal != original {
		t.Errorf("expected item value unchanged on rejection, got %q (was %q)", s.items[3].StringVal, original)
	}
	if !strings.Contains(s.editError, "invalid regex") {
		t.Errorf("expected editError to mention invalid regex, got %q", s.editError)
	}
}

func TestSettingsOverlay_ConfirmEdit_ValidRegex(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 3) // worktree.promptBranchPattern
	s.StartEditing()
	for s.editBuffer != "" {
		s.DeleteEditChar()
	}
	for _, ch := range "^release/.*$" {
		s.AddEditChar(ch)
	}

	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("ConfirmEdit unexpected error: %v", err)
	}
	if s.IsEditing() {
		t.Error("expected editor to close after valid input")
	}
	if s.items[3].StringVal != "^release/.*$" {
		t.Errorf("expected pattern saved, got %q", s.items[3].StringVal)
	}
}

func TestSettingsOverlay_ConfirmEdit_EmptyPatternIsValid(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	moveTo(s, 3) // worktree.promptBranchPattern
	s.StartEditing()
	for s.editBuffer != "" {
		s.DeleteEditChar()
	}

	if err := s.ConfirmEdit(); err != nil {
		t.Fatalf("ConfirmEdit on empty pattern returned error: %v", err)
	}
	if s.items[3].StringVal != "" {
		t.Errorf("expected empty pattern saved, got %q", s.items[3].StringVal)
	}
}

func TestSettingsOverlay_GHError(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	if s.HasGHError() {
		t.Fatal("should not have GH error initially")
	}

	s.SetGHError("gh not found")
	if !s.HasGHError() {
		t.Fatal("should have GH error after SetGHError")
	}

	s.DismissGHError()
	if s.HasGHError() {
		t.Fatal("should not have GH error after dismiss")
	}
}

func TestSettingsOverlay_Render(t *testing.T) {
	s := NewSettingsOverlay()
	cfg := &config.Config{
		Worktree: config.WorktreeConfig{Setup: "npm install"},
		OnComplete: config.OnCompleteConfig{
			Push:     true,
			CreatePR: false,
		},
	}
	s.LoadFromConfig(cfg)
	s.SetSize(80, 24)

	rendered := s.Render()

	if !strings.Contains(rendered, "Settings") {
		t.Error("expected 'Settings' in header")
	}
	if !strings.Contains(rendered, ".chief/config.yaml") {
		t.Error("expected config path in header")
	}

	for _, section := range []string{"Agent", "Worktree", "Bash", "On Complete"} {
		if !strings.Contains(rendered, section) {
			t.Errorf("expected %q section", section)
		}
	}

	if !strings.Contains(rendered, "npm install") {
		t.Error("expected 'npm install' value")
	}
	if !strings.Contains(rendered, "Yes") {
		t.Error("expected 'Yes' for push")
	}
	if !strings.Contains(rendered, "No") {
		t.Error("expected 'No' for createPR")
	}

	if !strings.Contains(rendered, "Esc: close") {
		t.Error("expected 'Esc: close' in footer")
	}
}

func TestSettingsOverlay_RenderEditError(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	moveTo(s, 3) // worktree.promptBranchPattern
	s.StartEditing()
	for s.editBuffer != "" {
		s.DeleteEditChar()
	}
	for _, ch := range "[bad" {
		s.AddEditChar(ch)
	}
	_ = s.ConfirmEdit()

	rendered := s.Render()
	if !strings.Contains(rendered, "invalid regex") {
		t.Error("expected rendered output to contain 'invalid regex' message")
	}
}

func TestSettingsOverlay_RenderGHError(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	s.SetGHError("gh not found")
	rendered := s.Render()

	if !strings.Contains(rendered, "GitHub CLI Error") {
		t.Error("expected 'GitHub CLI Error' in rendered output")
	}
	if !strings.Contains(rendered, "gh not found") {
		t.Error("expected error message in rendered output")
	}
	if !strings.Contains(rendered, "Press any key to dismiss") {
		t.Error("expected dismiss hint in footer")
	}
}

func TestSettingsOverlay_RenderEditing(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	s.StartEditing()
	rendered := s.Render()

	if !strings.Contains(rendered, "Enter: save") {
		t.Error("expected 'Enter: save' in footer during editing")
	}
}

func TestSettingsOverlay_RenderSelectedIndicator(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	rendered := s.Render()

	if !strings.Contains(rendered, ">") {
		t.Error("expected '>' cursor indicator for selected item")
	}
}

func TestSettingsOverlay_RenderEmptyStringValue(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())
	s.SetSize(80, 24)

	rendered := s.Render()

	if !strings.Contains(rendered, "(not set)") {
		t.Error("expected '(not set)' for empty setup command")
	}
}

func TestSettingsOverlay_GetSelectedItem(t *testing.T) {
	s := NewSettingsOverlay()
	s.LoadFromConfig(config.Default())

	item := s.GetSelectedItem()
	if item == nil {
		t.Fatal("expected non-nil selected item")
	}
	if item.Key != "agent.watchdogTimeout" {
		t.Errorf("expected first item key='agent.watchdogTimeout', got '%s'", item.Key)
	}

	s.MoveDown()
	item = s.GetSelectedItem()
	if item.Key != "worktree.setup" {
		t.Errorf("expected second item key='worktree.setup', got '%s'", item.Key)
	}
}
