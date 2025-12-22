package main

import (
	"context"
	"net/url"
	"time"

	"fyne.io/fyne/v2/app"

	"openrouter-costs-tray/internal/cache"
	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/logging"
	"openrouter-costs-tray/internal/notify"
	"openrouter-costs-tray/internal/openrouter"
	"openrouter-costs-tray/internal/refresh"
	"openrouter-costs-tray/internal/scheduler"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/ui/settings"
	"openrouter-costs-tray/internal/ui/tray"
)

const appID = "openrouter-costs-tray"

func main() {
	cfgPath, cfgErr := config.DefaultConfigPath()
	if cfgErr != nil {
		cfgPath = "config.json"
	}
	cachePath, cacheErr := cache.DefaultCachePath()
	if cacheErr != nil {
		cachePath = "costs_cache.json"
	}

	cfg, err := config.LoadFromPath(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	logger, levelVar := logging.NewLogger(cfg.Logging.Level)
	logger = logger.With("app", appID)

	if cfgErr != nil {
		logger.Warn("config dir unavailable", "error", cfgErr, "path", cfgPath)
	}
	if err != nil {
		logger.Warn("config load failed", "error", err, "path", cfgPath)
	} else {
		logger.Info("config loaded", "path", cfgPath)
	}

	if cacheErr != nil {
		logger.Warn("cache dir unavailable", "error", cacheErr, "path", cachePath)
	}

	cfgStore := config.NewStore(cfgPath, cfg)
	cacheStore := cache.NewStore(cachePath)

	stateStore := state.New()
	if cached, err := cacheStore.Load(); err == nil && cached != nil {
		logger.Info("cache loaded", "path", cachePath, "last_success_at", cached.LastSuccessAt)
		stateStore.SetSuccess(openrouter.Usage{
			Total:   cached.TotalUsage,
			Daily:   cached.DailyUsage,
			Weekly:  cached.WeeklyUsage,
			Monthly: cached.MonthlyUsage,
			KeyID:   cached.KeyID,
		}, cached.LastSuccessAt)
	} else if err != nil {
		logger.Warn("failed to load cache", "error", err)
	}

	if cfg.Connection.Token == "" {
		stateStore.SetNotConfigured()
	}

	fyneApp := app.NewWithID(appID)
	fyneApp.SetIcon(tray.IconResource())

	client := openrouter.NewClient("", nil, logger.With("component", "client"))
	notifier := notify.New(fyneApp, cfg.Notifications, logger.With("component", "notifier"))

	refresher := refresh.New(client, cacheStore, cfgStore, notifier, stateStore, logger.With("component", "refresher"))

	interval, ok := config.ParsePeriod(cfg.Updates.Period)
	if !ok {
		interval = 30 * time.Minute
	}
	sched := scheduler.New(interval, func(ctx context.Context) error {
		if err := refresher.Refresh(ctx); err != nil && err != refresh.ErrNotConfigured {
			return err
		}
		return nil
	}, logger.With("component", "scheduler"))

	var trayUI *tray.Tray
	trayActions := tray.Actions{
		Refresh: func() {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				if err := refresher.Refresh(ctx); err != nil && err != refresh.ErrNotConfigured {
					logger.Warn("manual refresh failed", "error", err)
				}
				cancel()
			}()
		},
		OpenSettings: func() {
			settings.Show(fyneApp, settings.Deps{
				ConfigStore: cfgStore,
				Refresher:   refresher,
				Scheduler:   sched,
				Notifier:    notifier,
				LevelVar:    levelVar,
				Logger:      logger.With("component", "settings"),
				OnConfigApplied: func(cfg config.Config) {
					trayUI.Update()
				},
			})
		},
		OpenWeb: func() {
			snap := stateStore.Snapshot()
			urlStr := activityURL(snap.Usage.KeyID)
			u, err := url.Parse(urlStr)
			if err != nil {
				logger.Warn("invalid activity url", "error", err)
				return
			}
			if err := fyneApp.OpenURL(u); err != nil {
				logger.Warn("open url failed", "error", err)
			}
		},
		Exit: func() {
			sched.Stop()
			fyneApp.Quit()
		},
	}

	trayUI = tray.New(fyneApp, stateStore, cfgStore, logger.With("component", "tray"), trayActions)

	refresher.SetUpdateCallback(func() {
		trayUI.Update()
	})

	trayUI.Update()
	sched.Start()

	if cfg.Updates.UpdateOnStart {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			if err := refresher.Refresh(ctx); err != nil && err != refresh.ErrNotConfigured {
				logger.Warn("startup refresh failed", "error", err)
			}
			cancel()
		}()
	}

	fyneApp.Run()
}

func activityURL(keyID string) string {
	if keyID == "" {
		return "https://openrouter.ai/activity"
	}
	return "https://openrouter.ai/activity?api_key_id=" + url.QueryEscape(keyID)
}
