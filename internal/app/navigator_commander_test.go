package app

import (
	"strings"
	"testing"

	"github.com/Gaurav-Gosain/tuios/internal/config"
	"github.com/charmbracelet/x/ansi"
)

// TestNavigatorCommanderToggle tests activating and toggling commander mode.
func TestNavigatorCommanderToggle(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	if m.NavigatorCommander() {
		t.Fatal("expected commander mode to be off initially")
	}
	if m.NavigatorLayout() != config.NavigatorLayoutTree {
		t.Fatalf("expected initial layout %q, got %q", config.NavigatorLayoutTree, m.NavigatorLayout())
	}

	// Toggle on
	m.NavigatorToggleCommander()
	if !m.NavigatorCommander() {
		t.Fatal("expected commander mode to be active after toggle")
	}
	if m.NavigatorLayout() != config.NavigatorLayoutCommander {
		t.Fatalf("expected layout %q, got %q", config.NavigatorLayoutCommander, m.NavigatorLayout())
	}

	// Toggle off
	m.NavigatorToggleCommander()
	if m.NavigatorCommander() {
		t.Fatal("expected commander mode to be inactive after second toggle")
	}
	if m.NavigatorLayout() != config.NavigatorLayoutTree {
		t.Fatalf("expected layout %q, got %q", config.NavigatorLayoutTree, m.NavigatorLayout())
	}
}

// TestNavigatorCommanderSwitchSideAndNavigation tests independent dual-panel navigation.
func TestNavigatorCommanderSwitchSideAndNavigation(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	// Seed test sessions
	m.navigator.sessions = []navSession{
		{
			Name:    "alpha",
			Current: true,
			Panes: []navPane{
				{ID: "p1", Title: "editor", Workspace: 1},
				{ID: "p2", Title: "terminal", Workspace: 1},
				{ID: "p3", Title: "server", Workspace: 2},
			},
		},
		{
			Name: "beta",
			Panes: []navPane{
				{ID: "p4", Title: "db", Workspace: 1},
			},
		},
	}

	m.NavigatorToggleCommander()
	if m.NavigatorActiveSide() != 0 {
		t.Fatalf("expected initial active side 0 (Left), got %d", m.NavigatorActiveSide())
	}

	// Move cursor on Left panel
	m.NavigatorSelect(1)
	if m.navActiveList().Cursor != 1 {
		t.Fatalf("expected left cursor 1, got %d", m.navActiveList().Cursor)
	}

	// Switch to Right panel
	m.NavigatorSwitchSide()
	if m.NavigatorActiveSide() != 1 {
		t.Fatalf("expected active side 1 (Right), got %d", m.NavigatorActiveSide())
	}

	// Move cursor on Right panel independently
	m.NavigatorSelect(2)
	if m.navActiveList().Cursor != 2 {
		t.Fatalf("expected right cursor 2, got %d", m.navActiveList().Cursor)
	}

	// Switch back to Left panel and verify left cursor was preserved
	m.NavigatorSwitchSide()
	if m.NavigatorActiveSide() != 0 {
		t.Fatalf("expected active side 0 (Left), got %d", m.NavigatorActiveSide())
	}
	if m.navActiveList().Cursor != 1 {
		t.Fatalf("expected left cursor still 1, got %d", m.navActiveList().Cursor)
	}
}

// TestNavigatorCommanderRendering verifies the dual-panel visual presentation.
func TestNavigatorCommanderRendering(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	m.navigator.sessions = []navSession{
		{
			Name:    "main",
			Current: true,
			Panes: []navPane{
				{ID: "win1", Title: "bash", Workspace: 1},
			},
		},
	}

	m.NavigatorToggleCommander()
	content, _, _ := m.renderNavigator()
	plain := ansi.Strip(content)

	if !strings.Contains(plain, "Pane Commander") {
		t.Errorf("rendered panel missing 'Pane Commander' title: %s", plain)
	}
	if !strings.Contains(plain, "Left") || !strings.Contains(plain, "Right") {
		t.Errorf("rendered panel missing dual 'Left' and 'Right' boxes: %s", plain)
	}
	if !strings.Contains(plain, "Tab") || !strings.Contains(plain, "move") || !strings.Contains(plain, "normal") {
		t.Errorf("rendered panel missing commander hints: %s", plain)
	}
}

// TestNavigatorCommanderPromptNew verifies the 'n' quick-create prompt state.
func TestNavigatorCommanderPromptNew(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	m.NavigatorToggleCommander()
	m.NavigatorSetCmdPrompt("new")

	if m.NavigatorCmdPrompt() != "new" {
		t.Fatalf("expected prompt 'new', got %q", m.NavigatorCmdPrompt())
	}
	if !strings.Contains(m.NavigatorCmdNotice(), "Create: [s] Session") {
		t.Fatalf("unexpected prompt notice: %q", m.NavigatorCmdNotice())
	}

	// Cancel prompt
	m.NavigatorSetCmdPrompt("")
	if m.NavigatorCmdPrompt() != "" {
		t.Fatalf("expected prompt cleared, got %q", m.NavigatorCmdPrompt())
	}
}

// TestNavigatorCommanderMoveItemValidation verifies move logic guards.
func TestNavigatorCommanderMoveItemValidation(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	m.navigator.sessions = []navSession{
		{
			Name:    "main",
			Current: true,
			Panes: []navPane{
				{ID: "win1", Title: "bash", Workspace: 1},
			},
		},
	}

	m.NavigatorToggleCommander()
	rows := m.navigatorRows()
	if len(rows) == 0 {
		t.Fatal("expected navigator rows")
	}

	// Cursor on Session row (index 0)
	m.NavigatorSelect(0)
	m.navigator.rightList.Cursor = 0

	cmd := m.NavigatorMoveItem()
	if cmd != nil {
		t.Fatal("expected move on session row to return nil cmd")
	}
	if !strings.Contains(m.NavigatorCmdNotice(), "Cannot move a whole session") {
		t.Fatalf("unexpected notice for session move: %q", m.NavigatorCmdNotice())
	}
}

// TestNavigatorAllSessionsExpandedByDefault verifies that both current and non-current sessions
// (and their workspaces and panes) are expanded and visible by default.
func TestNavigatorAllSessionsExpandedByDefault(t *testing.T) {
	m := newNarrowOS(t, 120, 40)
	m.OpenNavigator()
	defer m.CloseNavigator()

	m.navigator.sessions = []navSession{
		{
			Name: "foreign1",
			Panes: []navPane{
				{ID: "f1p1", Title: "worker1", Workspace: 1},
				{ID: "f1p2", Title: "worker2", Workspace: 2},
			},
		},
		{
			Name:    "current_sess",
			Current: true,
			Panes: []navPane{
				{ID: "cp1", Title: "main_terminal", Workspace: 1},
			},
		},
		{
			Name: "foreign2",
			WorkspaceNames: map[int]string{3: "empty_ws"},
			Panes: []navPane{
				{ID: "f2p1", Title: "logs", Workspace: 1},
			},
		},
	}

	rows := m.navigatorRows()
	if len(rows) != 12 {
		t.Fatalf("expected 12 rows with all sessions expanded, got %d", len(rows))
	}

	// Verify foreign1 pane is in rows
	foundF1P1 := false
	for _, r := range rows {
		if r.Kind == navRowPane && m.navigator.sessions[r.Session].Panes[r.Pane].ID == "f1p1" {
			foundF1P1 = true
			break
		}
	}
	if !foundF1P1 {
		t.Fatal("expected pane f1p1 from non-current session to be visible")
	}

	// Verify empty named workspace 3 from foreign2 is in rows
	foundWS3 := false
	for _, r := range rows {
		if r.Kind == navRowWorkspace && r.Workspace == 3 && m.navigator.sessions[r.Session].Name == "foreign2" {
			foundWS3 = true
			break
		}
	}
	if !foundWS3 {
		t.Fatal("expected empty named workspace 3 from foreign2 to be visible")
	}

	// Fold foreign1 (row 0)
	m.NavigatorSelect(0)
	m.NavigatorToggle()
	rowsAfterFold := m.navigatorRows()
	if len(rowsAfterFold) != 8 {
		t.Fatalf("expected 8 rows after folding foreign1, got %d", len(rowsAfterFold))
	}

	// Unfold foreign1
	m.NavigatorToggle()
	rowsAfterUnfold := m.navigatorRows()
	if len(rowsAfterUnfold) != 12 {
		t.Fatalf("expected 12 rows after unfolding foreign1, got %d", len(rowsAfterUnfold))
	}
}
