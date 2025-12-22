package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Scheduler struct {
	mu       sync.Mutex
	interval time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
	running  bool
	refresh  func(context.Context) error
	logger   *slog.Logger
}

func New(interval time.Duration, refresh func(context.Context) error, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{interval: interval, refresh: refresh, logger: logger}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.ticker = time.NewTicker(s.interval)
	s.logger.Info("scheduler started", "interval", s.interval)
	stopCh := s.stopCh
	ticker := s.ticker
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				if err := s.refresh(ctx); err != nil {
					s.logger.Warn("scheduled refresh failed", "error", err)
				}
				cancel()
			case <-stopCh:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.stopCh != nil {
		close(s.stopCh)
	}
	s.ticker = nil
	s.stopCh = nil
	s.mu.Unlock()
	s.logger.Info("scheduler stopped")
}

func (s *Scheduler) Reschedule(interval time.Duration) {
	if interval <= 0 {
		return
	}
	s.mu.Lock()
	s.interval = interval
	wasRunning := s.running
	s.mu.Unlock()
	if wasRunning {
		s.Stop()
		s.Start()
	}
	s.logger.Info("scheduler rescheduled", "interval", interval)
}

func (s *Scheduler) Interval() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.interval
}
