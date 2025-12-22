package notify

import (
	"log/slog"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/util"
)

const errorNotifyInterval = 10 * time.Minute

type Notifier struct {
	app         fyne.App
	logger      *slog.Logger
	mu          sync.RWMutex
	cfg         config.NotificationsConfig
	lastErrorAt time.Time
}

func New(app fyne.App, cfg config.NotificationsConfig, logger *slog.Logger) *Notifier {
	if logger == nil {
		logger = slog.Default()
	}
	return &Notifier{app: app, cfg: cfg, logger: logger}
}

func (n *Notifier) UpdateConfig(cfg config.NotificationsConfig) {
	n.mu.Lock()
	n.cfg = cfg
	n.mu.Unlock()
}

func (n *Notifier) NotifyUpdateSpent(amount float64) {
	n.mu.RLock()
	cfg := n.cfg
	n.mu.RUnlock()
	if !cfg.Enabled || !cfg.OnUpdateSpent {
		return
	}
	msg := "Recently spent: " + util.FormatUSD(amount)
	n.send("OpenRouter Costs", msg)
}

func (n *Notifier) NotifyError(err error) {
	if err == nil {
		return
	}
	n.mu.Lock()
	cfg := n.cfg
	if !cfg.Enabled || !cfg.OnError {
		n.mu.Unlock()
		return
	}
	if time.Since(n.lastErrorAt) < errorNotifyInterval {
		n.mu.Unlock()
		return
	}
	n.lastErrorAt = time.Now()
	n.mu.Unlock()

	n.send("OpenRouter Costs", "Error: "+err.Error()+" (retrying on schedule)")
}

func (n *Notifier) send(title, content string) {
	if n.app == nil {
		n.logger.Warn("notification dropped: no app", "title", title, "content", content)
		return
	}
	n.app.SendNotification(&fyne.Notification{Title: title, Content: content})
}
