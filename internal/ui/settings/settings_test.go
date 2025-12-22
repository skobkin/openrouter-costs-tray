package settings

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"openrouter-costs-tray/internal/config"
)

func TestNotificationChecksToggle(t *testing.T) {
	app := test.NewApp()
	window = nil

	cfg := config.DefaultConfig()
	cfg.Notifications.Enabled = false
	cfg.Notifications.OnUpdateSpent = true
	cfg.Notifications.OnError = true
	cfg.Notifications.OnStartSummary = true
	store := config.NewStore("unused", cfg)

	Show(app, Deps{ConfigStore: store})
	defer func() {
		if window != nil {
			window.Close()
			window = nil
		}
	}()

	checks := map[string]*widget.Check{}
	collectChecks(window.Content(), checks)

	enable := checks["Enable notifications"]
	update := checks["On update: spent"]
	errorCheck := checks["On error"]
	startSummary := checks["On start: spends summary"]
	if enable == nil || update == nil || errorCheck == nil || startSummary == nil {
		t.Fatalf("expected notification checks to exist")
	}

	if enable.Checked {
		t.Fatalf("expected notifications disabled by default")
	}
	if !update.Disabled() || !errorCheck.Disabled() || !startSummary.Disabled() {
		t.Fatalf("expected notification options disabled when notifications off")
	}

	enable.SetChecked(true)
	if update.Disabled() || errorCheck.Disabled() || startSummary.Disabled() {
		t.Fatalf("expected notification options enabled after toggle")
	}
}

func collectChecks(obj fyne.CanvasObject, out map[string]*widget.Check) {
	if obj == nil {
		return
	}
	if check, ok := obj.(*widget.Check); ok {
		out[check.Text] = check
	}
	if container, ok := obj.(*fyne.Container); ok {
		for _, child := range container.Objects {
			collectChecks(child, out)
		}
	}
}
