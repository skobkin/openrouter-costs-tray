package tray

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/systray"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/summary"
)

type Actions struct {
	Refresh      func()
	OpenSettings func()
	OpenWeb      func()
	Exit         func()
}

type Tray struct {
	app        fyne.App
	desktopApp desktop.App
	state      *state.State
	cfgStore   *config.Store
	logger     *slog.Logger
	menu       *fyne.Menu
	actions    Actions
}

func New(app fyne.App, stateStore *state.State, cfgStore *config.Store, logger *slog.Logger, actions Actions) *Tray {
	if logger == nil {
		logger = slog.Default()
	}
	if actions.Refresh == nil {
		actions.Refresh = func() {}
	}
	if actions.OpenSettings == nil {
		actions.OpenSettings = func() {}
	}
	if actions.OpenWeb == nil {
		actions.OpenWeb = func() {}
	}
	if actions.Exit == nil {
		actions.Exit = func() {}
	}

	var desktopApp desktop.App
	if app != nil {
		if desk, ok := app.(desktop.App); ok {
			desktopApp = desk
		} else if logger != nil {
			logger.Warn("system tray not supported by app")
		}
	}
	tray := &Tray{
		app:        app,
		desktopApp: desktopApp,
		state:      stateStore,
		cfgStore:   cfgStore,
		logger:     logger,
		actions:    actions,
	}
	tray.buildMenu()
	tray.Update()
	return tray
}

func (t *Tray) Update() {
	if t.app == nil || t.desktopApp == nil {
		return
	}
	cfg := t.cfgStore.Get()
	snap := t.state.Snapshot()
	label := summary.Tooltip(cfg, snap)

	t.menu.Label = label
	t.setIcon(snap, cfg)
	t.menu.Refresh()
	systray.SetTooltip(label)
}

func (t *Tray) buildMenu() {
	refreshItem := fyne.NewMenuItem("Refresh", func() {
		t.actions.Refresh()
	})
	settingsItem := fyne.NewMenuItem("Settings", func() {
		t.actions.OpenSettings()
	})
	openWebItem := fyne.NewMenuItem("Open in web", func() {
		t.actions.OpenWeb()
	})
	exitItem := fyne.NewMenuItem("Exit", func() {
		t.actions.Exit()
	})

	t.menu = fyne.NewMenu("OpenRouter Costs", refreshItem, openWebItem, settingsItem, exitItem)
	if t.desktopApp != nil {
		t.desktopApp.SetSystemTrayMenu(t.menu)
	}
}

func (t *Tray) setIcon(snap state.Snapshot, cfg config.Config) {
	if t.desktopApp == nil {
		return
	}
	if snap.NotConfigured || cfg.Connection.Token == "" {
		t.desktopApp.SetSystemTrayIcon(IconResource())
		return
	}
	if snap.LastError != "" {
		t.desktopApp.SetSystemTrayIcon(theme.ErrorIcon())
		return
	}
	t.desktopApp.SetSystemTrayIcon(IconResource())
}
