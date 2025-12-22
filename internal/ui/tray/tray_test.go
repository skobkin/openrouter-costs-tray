package tray

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/state"
)

type stubDesktopApp struct {
	lastMenu *fyne.Menu
	lastIcon fyne.Resource
}

func (s *stubDesktopApp) SetSystemTrayMenu(menu *fyne.Menu) {
	s.lastMenu = menu
}

func (s *stubDesktopApp) SetSystemTrayIcon(icon fyne.Resource) {
	s.lastIcon = icon
}

func TestBuildMenu(t *testing.T) {
	tr := &Tray{}
	tr.buildMenu()
	if tr.menu == nil {
		t.Fatalf("expected menu to be built")
	}
	if tr.menu.Label != "OpenRouter Costs" {
		t.Fatalf("unexpected menu label: %s", tr.menu.Label)
	}
	if len(tr.menu.Items) != 4 {
		t.Fatalf("expected 4 menu items, got %d", len(tr.menu.Items))
	}
	labels := []string{"Refresh", "Open in web", "Settings", "Exit"}
	for i, label := range labels {
		if tr.menu.Items[i].Label != label {
			t.Fatalf("expected item %d label %q, got %q", i, label, tr.menu.Items[i].Label)
		}
	}
}

func TestSetIcon(t *testing.T) {
	stub := &stubDesktopApp{}
	tr := &Tray{desktopApp: stub}

	cfg := config.DefaultConfig()
	cfg.Connection.Token = ""
	tr.setIcon(state.Snapshot{NotConfigured: true}, cfg)
	if stub.lastIcon == nil || stub.lastIcon.Name() != IconResource().Name() {
		t.Fatalf("expected tray icon for not configured")
	}

	cfg.Connection.Token = "token"
	tr.setIcon(state.Snapshot{LastError: "boom"}, cfg)
	if stub.lastIcon == nil || stub.lastIcon.Name() != theme.ErrorIcon().Name() {
		t.Fatalf("expected error icon")
	}

	tr.setIcon(state.Snapshot{}, cfg)
	if stub.lastIcon == nil || stub.lastIcon.Name() != IconResource().Name() {
		t.Fatalf("expected tray icon for success")
	}
}
