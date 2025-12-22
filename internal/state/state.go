package state

import (
	"sync"
	"time"

	"openrouter-costs-tray/internal/openrouter"
)

type Snapshot struct {
	LastSuccessAt time.Time
	Usage         openrouter.Usage
	LastError     string
	NotConfigured bool
}

type State struct {
	mu            sync.RWMutex
	lastSuccessAt time.Time
	usage         openrouter.Usage
	lastError     string
	notConfigured bool
}

func New() *State {
	return &State{}
}

func (s *State) SetNotConfigured() {
	s.mu.Lock()
	s.notConfigured = true
	s.lastError = ""
	s.mu.Unlock()
}

func (s *State) ClearNotConfigured() {
	s.mu.Lock()
	s.notConfigured = false
	s.mu.Unlock()
}

func (s *State) SetSuccess(usage openrouter.Usage, at time.Time) {
	s.mu.Lock()
	s.notConfigured = false
	s.usage = usage
	s.lastSuccessAt = at
	s.lastError = ""
	s.mu.Unlock()
}

func (s *State) SetError(err error) {
	s.mu.Lock()
	s.notConfigured = false
	if err != nil {
		s.lastError = err.Error()
	}
	s.mu.Unlock()
}

func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{
		LastSuccessAt: s.lastSuccessAt,
		Usage:         s.usage,
		LastError:     s.lastError,
		NotConfigured: s.notConfigured,
	}
}
