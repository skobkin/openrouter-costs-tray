package notify

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/util"
)

func TestNotifyUpdateSpentEnabled(t *testing.T) {
	app := test.NewApp()
	cfg := config.NotificationsConfig{Enabled: true, OnUpdateSpent: true}
	n := New(app, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	amount := 1.23
	expected := &fyne.Notification{
		Title:   "OpenRouter Costs",
		Content: "Recently spent: " + util.FormatUSD(amount),
	}
	test.AssertNotificationSent(t, expected, func() {
		n.NotifyUpdateSpent(amount)
	})
}

func TestNotifyUpdateSpentDisabled(t *testing.T) {
	app := test.NewApp()
	cfg := config.NotificationsConfig{Enabled: false, OnUpdateSpent: true}
	n := New(app, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	test.AssertNotificationSent(t, nil, func() {
		n.NotifyUpdateSpent(1.23)
	})
}

func TestNotifyErrorThrottled(t *testing.T) {
	app := test.NewApp()
	cfg := config.NotificationsConfig{Enabled: true, OnError: true}
	n := New(app, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	err := errors.New("boom")
	first := &fyne.Notification{
		Title:   "OpenRouter Costs",
		Content: "Error: " + err.Error() + " (retrying on schedule)",
	}
	test.AssertNotificationSent(t, first, func() {
		n.NotifyError(err)
	})

	test.AssertNotificationSent(t, nil, func() {
		n.NotifyError(err)
	})
}

func TestNotifyStartSummary(t *testing.T) {
	app := test.NewApp()
	cfg := config.NotificationsConfig{Enabled: true, OnStartSummary: true}
	n := New(app, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	expected := &fyne.Notification{
		Title:   "OpenRouter Costs",
		Content: "Summary",
	}
	test.AssertNotificationSent(t, expected, func() {
		n.NotifyStartSummary("Summary")
	})
}

func TestNotifyErrorRespectsInterval(t *testing.T) {
	app := test.NewApp()
	cfg := config.NotificationsConfig{Enabled: true, OnError: true}
	n := New(app, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	n.lastErrorAt = time.Now().Add(-errorNotifyInterval)
	first := &fyne.Notification{
		Title:   "OpenRouter Costs",
		Content: "Error: boom (retrying on schedule)",
	}
	test.AssertNotificationSent(t, first, func() {
		n.NotifyError(errors.New("boom"))
	})
}
