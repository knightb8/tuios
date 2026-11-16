package input

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/Gaurav-Gosain/tuios/internal/app"
	"github.com/Gaurav-Gosain/tuios/internal/config"
)

func newTestNavigatorOS() *app.OS {
	m := app.NewOS(app.OSOptions{UserConfig: config.DefaultConfig()})
	if m.KeybindRegistry == nil {
		m.KeybindRegistry = config.NewKeybindRegistry(config.DefaultConfig())
	}
	m.Width, m.Height = 120, 40
	m.EffectiveWidth, m.EffectiveHeight = 120, 40
	m.OpenNavigator()
	return m
}

func testKey(key string) tea.KeyPressMsg {
	switch key {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	default:
		runes := []rune(key)
		return tea.KeyPressMsg{Code: runes[0], Text: key}
	}
}

// TestNavigatorCommanderInputKeys tests keyboard shortcuts for Commander mode.
func TestNavigatorCommanderInputKeys(t *testing.T) {
	o := newTestNavigatorOS()
	defer o.CloseNavigator()

	if o.NavigatorCommander() {
		t.Fatal("expected commander mode off")
	}

	// Press 'c' to enter Commander mode
	o, _ = handleNavigatorInput(testKey("c"), o)
	if !o.NavigatorCommander() {
		t.Fatal("expected 'c' to toggle Commander mode on")
	}

	// Press 'tab' to switch to Right panel
	if o.NavigatorActiveSide() != 0 {
		t.Fatalf("expected active side 0, got %d", o.NavigatorActiveSide())
	}
	o, _ = handleNavigatorInput(testKey("tab"), o)
	if o.NavigatorActiveSide() != 1 {
		t.Fatalf("expected 'tab' to switch active side to 1, got %d", o.NavigatorActiveSide())
	}
	o, _ = handleNavigatorInput(testKey("tab"), o)
	if o.NavigatorActiveSide() != 0 {
		t.Fatalf("expected 'tab' to switch active side back to 0, got %d", o.NavigatorActiveSide())
	}

	// Press 'n' to open quick-create prompt
	o, _ = handleNavigatorInput(testKey("n"), o)
	if o.NavigatorCmdPrompt() != "new" {
		t.Fatalf("expected 'n' to set prompt to 'new', got %q", o.NavigatorCmdPrompt())
	}

	// Press 'esc' to cancel quick-create prompt
	o, _ = handleNavigatorInput(testKey("esc"), o)
	if o.NavigatorCmdPrompt() != "" {
		t.Fatalf("expected 'esc' to cancel prompt, got %q", o.NavigatorCmdPrompt())
	}

	// Press 'c' to leave Commander mode
	o, _ = handleNavigatorInput(testKey("c"), o)
	if o.NavigatorCommander() {
		t.Fatal("expected 'c' to toggle Commander mode off")
	}
}
