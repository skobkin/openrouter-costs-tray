package refresh

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"openrouter-costs-tray/internal/cache"
	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/notify"
	"openrouter-costs-tray/internal/openrouter"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/util"
)

var ErrNotConfigured = errors.New("token not configured")

// Refresher handles fetching usage, updating cache/state, and notifying.
type Refresher struct {
	client   *openrouter.Client
	cache    *cache.Store
	config   *config.Store
	notifier *notify.Notifier
	state    *state.State
	logger   *slog.Logger
	updateFn func()
}

func New(client *openrouter.Client, cacheStore *cache.Store, cfgStore *config.Store, notifier *notify.Notifier, stateStore *state.State, logger *slog.Logger) *Refresher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Refresher{
		client:   client,
		cache:    cacheStore,
		config:   cfgStore,
		notifier: notifier,
		state:    stateStore,
		logger:   logger,
	}
}

func (r *Refresher) SetUpdateCallback(fn func()) {
	r.updateFn = fn
}

func computeDelta(prev *cache.CostsCache, tokenHash string, currentTotal float64) float64 {
	if prev == nil || prev.KeyHash == "" || prev.KeyHash != tokenHash {
		return 0
	}
	delta := currentTotal - prev.TotalUsage
	if delta < 0 {
		return 0
	}
	return delta
}

func (r *Refresher) Refresh(ctx context.Context) error {
	cfg := r.config.Get()
	token := cfg.Connection.Token
	if token == "" {
		r.logger.Info("refresh skipped: not configured")
		r.state.SetNotConfigured()
		r.triggerUpdate()
		return ErrNotConfigured
	}
	r.state.ClearNotConfigured()

	r.logger.Info("refresh started")
	usage, err := r.client.FetchUsage(ctx, token)
	if err != nil {
		r.logger.Error("refresh failed", "error", err)
		r.state.SetError(err)
		if r.notifier != nil {
			r.notifier.NotifyError(err)
		}
		r.triggerUpdate()
		return err
	}

	var lastCache *cache.CostsCache

	if r.cache != nil {
		cached, err := r.cache.Load()
		if err != nil {
			r.logger.Warn("cache load failed", "error", err)
		} else {
			lastCache = cached
		}
	}

	tokenHash := util.TokenHash(token)
	delta := computeDelta(lastCache, tokenHash, usage.Total)
	r.logger.Info("usage delta computed", "delta", delta)
	if lastCache != nil && lastCache.KeyHash == tokenHash && delta == 0 && usage.Total < lastCache.TotalUsage {
		r.logger.Warn("usage total decreased", "previous", lastCache.TotalUsage, "current", usage.Total)
	}

	now := time.Now().UTC()
	newCache := cache.CostsCache{
		SchemaVersion: cache.SchemaVersion,
		LastSuccessAt: now,
		TotalUsage:    usage.Total,
		DailyUsage:    usage.Daily,
		WeeklyUsage:   usage.Weekly,
		MonthlyUsage:  usage.Monthly,
		KeyHash:       tokenHash,
		KeyID:         usage.KeyID,
	}

	r.logger.Info("refresh succeeded", "total", usage.Total)

	if r.cache != nil {
		if err := r.cache.Save(newCache); err != nil {
			r.logger.Warn("cache save failed", "error", err)
		}
	}

	r.state.SetSuccess(usage, now)
	if delta > 0 {
		if r.notifier != nil {
			r.notifier.NotifyUpdateSpent(delta)
		}
	}
	r.triggerUpdate()
	return nil
}

func (r *Refresher) TestToken(ctx context.Context, token string) (openrouter.Usage, error) {
	return r.client.FetchUsage(ctx, token)
}

func (r *Refresher) triggerUpdate() {
	if r.updateFn != nil {
		r.updateFn()
	}
}
