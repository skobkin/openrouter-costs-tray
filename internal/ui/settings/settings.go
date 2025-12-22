package settings

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/logging"
	"openrouter-costs-tray/internal/notify"
	"openrouter-costs-tray/internal/refresh"
	"openrouter-costs-tray/internal/scheduler"
)

type Deps struct {
	ConfigStore     *config.Store
	Refresher       *refresh.Refresher
	Scheduler       *scheduler.Scheduler
	Notifier        *notify.Notifier
	LevelVar        *slog.LevelVar
	Logger          *slog.Logger
	OnConfigApplied func(config.Config)
}

var window fyne.Window

func Show(app fyne.App, deps Deps) {
	if window != nil {
		window.Show()
		window.RequestFocus()
		return
	}
	settingsLogger := deps.Logger
	if settingsLogger == nil {
		settingsLogger = slog.Default()
	}

	window = app.NewWindow("Settings")
	window.Resize(fyne.NewSize(420, 420))

	cfg := deps.ConfigStore.Get()

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("OpenRouter API key")
	tokenEntry.SetText(cfg.Connection.Token)
	statusLabel := widget.NewLabel("")

	testButton := widget.NewButton("Test", func() {
		token := strings.TrimSpace(tokenEntry.Text)
		if token == "" {
			dialog.ShowInformation("Test", "Token is empty", window)
			return
		}
		statusLabel.SetText("Testing...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			usage, err := deps.Refresher.TestToken(ctx, token)
			cancel()
			result := "Test OK"
			if err != nil {
				result = "Test failed: " + err.Error()
			} else if usage.Label != "" {
				result = "Test OK (" + usage.Label + ")"
			}
			runOnMain(func() {
				statusLabel.SetText(result)
			})
		}()
	})

	connectionRow := container.NewBorder(nil, nil, nil, testButton, tokenEntry)

	periodSelect := widget.NewSelect(config.PeriodOptions, nil)
	periodSelect.SetSelected(cfg.Updates.Period)
	updateOnStart := widget.NewCheck("Update on start", nil)
	updateOnStart.SetChecked(cfg.Updates.UpdateOnStart)

	notifyEnabled := widget.NewCheck("Enable notifications", nil)
	notifyEnabled.SetChecked(cfg.Notifications.Enabled)
	notifyUpdate := widget.NewCheck("On update: spent", nil)
	notifyUpdate.SetChecked(cfg.Notifications.OnUpdateSpent)
	notifyError := widget.NewCheck("On error", nil)
	notifyError.SetChecked(cfg.Notifications.OnError)
	notifyStartSummary := widget.NewCheck("On start: spends summary", nil)
	notifyStartSummary.SetChecked(cfg.Notifications.OnStartSummary)
	testNotifyButton := widget.NewButton("Test notification", func() {
		app.SendNotification(&fyne.Notification{
			Title:   "OpenRouter Costs",
			Content: "Test notification",
		})
	})

	setNotificationsEnabled := func(enabled bool) {
		if enabled {
			notifyUpdate.Enable()
			notifyError.Enable()
			notifyStartSummary.Enable()
		} else {
			notifyUpdate.Disable()
			notifyError.Disable()
			notifyStartSummary.Disable()
		}
	}
	setNotificationsEnabled(cfg.Notifications.Enabled)
	notifyEnabled.OnChanged = func(value bool) {
		setNotificationsEnabled(value)
	}

	logLevelSelect := widget.NewSelect([]string{"debug", "info", "warn", "error"}, nil)
	logLevelSelect.SetSelected(cfg.Logging.Level)

	saveButton := widget.NewButton("Save", func() {
		newCfg := config.Config{
			Connection: config.ConnectionConfig{Token: strings.TrimSpace(tokenEntry.Text)},
			Updates: config.UpdatesConfig{
				Period:        periodSelect.Selected,
				UpdateOnStart: updateOnStart.Checked,
			},
			Notifications: config.NotificationsConfig{
				Enabled:        notifyEnabled.Checked,
				OnUpdateSpent:  notifyUpdate.Checked,
				OnError:        notifyError.Checked,
				OnStartSummary: notifyStartSummary.Checked,
			},
			Logging: config.LoggingConfig{Level: logLevelSelect.Selected},
		}
		config.Normalize(&newCfg)
		deps.ConfigStore.Set(newCfg)
		if err := deps.ConfigStore.Save(); err != nil {
			statusLabel.SetText("Save failed: " + err.Error())
			return
		}
		settingsLogger.Info("config saved")
		if deps.LevelVar != nil {
			logging.SetLevel(deps.LevelVar, newCfg.Logging.Level)
		}
		if deps.Notifier != nil {
			deps.Notifier.UpdateConfig(newCfg.Notifications)
		}
		if deps.Scheduler != nil {
			if dur, ok := config.ParsePeriod(newCfg.Updates.Period); ok {
				deps.Scheduler.Reschedule(dur)
			}
		}
		if deps.OnConfigApplied != nil {
			deps.OnConfigApplied(newCfg)
		}
		statusLabel.SetText("Saved. Updating...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			err := deps.Refresher.Refresh(ctx)
			cancel()
			result := "Saved + Updated OK"
			if err != nil {
				result = "Saved + Update failed: " + err.Error()
			}
			runOnMain(func() {
				statusLabel.SetText(result)
			})
		}()
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Connection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		connectionRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Update settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, widget.NewLabel("Period"), periodSelect),
		updateOnStart,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Notifications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		notifyEnabled,
		indentCheck(notifyUpdate),
		indentCheck(notifyStartSummary),
		indentCheck(notifyError),
		testNotifyButton,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Logging", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2, widget.NewLabel("Level"), logLevelSelect),
		layout.NewSpacer(),
		container.NewHBox(saveButton),
		statusLabel,
	)

	window.SetContent(container.NewPadded(form))
	window.SetOnClosed(func() {
		window = nil
	})
	window.Show()
}

func runOnMain(fn func()) {
	if fn == nil {
		return
	}
	fn()
}

func indentCheck(check *widget.Check) fyne.CanvasObject {
	return container.NewHBox(widget.NewLabel("    "), check)
}
