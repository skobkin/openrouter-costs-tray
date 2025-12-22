package state

import (
	"errors"
	"testing"
	"time"

	"openrouter-costs-tray/internal/openrouter"
)

func TestStateTransitions(t *testing.T) {
	s := New()
	snap := s.Snapshot()
	if snap.NotConfigured {
		t.Fatalf("expected not configured false by default")
	}
	if snap.LastError != "" {
		t.Fatalf("expected empty error by default")
	}

	s.SetNotConfigured()
	snap = s.Snapshot()
	if !snap.NotConfigured {
		t.Fatalf("expected not configured true")
	}
	if snap.LastError != "" {
		t.Fatalf("expected error cleared on not configured")
	}

	s.ClearNotConfigured()
	snap = s.Snapshot()
	if snap.NotConfigured {
		t.Fatalf("expected not configured cleared")
	}

	err := errors.New("boom")
	s.SetError(err)
	snap = s.Snapshot()
	if snap.LastError != err.Error() {
		t.Fatalf("expected error message set")
	}
	if snap.NotConfigured {
		t.Fatalf("expected not configured false after error")
	}

	usage := openrouter.Usage{Total: 10.5}
	when := time.Now().UTC()
	s.SetSuccess(usage, when)
	snap = s.Snapshot()
	if snap.LastError != "" {
		t.Fatalf("expected error cleared on success")
	}
	if snap.NotConfigured {
		t.Fatalf("expected not configured false on success")
	}
	if snap.Usage.Total != usage.Total {
		t.Fatalf("expected usage to be set")
	}
	if !snap.LastSuccessAt.Equal(when) {
		t.Fatalf("expected last success time set")
	}
}
