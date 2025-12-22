package refresh

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"openrouter-costs-tray/internal/cache"
	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/openrouter"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/util"
)

func TestComputeDelta(t *testing.T) {
	prev := &cache.CostsCache{TotalUsage: 10, KeyHash: "abc"}

	cases := []struct {
		name     string
		prev     *cache.CostsCache
		hash     string
		current  float64
		expected float64
	}{
		{"no prev", nil, "abc", 10, 0},
		{"hash mismatch", prev, "def", 12, 0},
		{"increase", prev, "abc", 12, 2},
		{"equal", prev, "abc", 10, 0},
		{"decrease", prev, "abc", 8, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeDelta(tc.prev, tc.hash, tc.current)
			if got != tc.expected {
				t.Fatalf("expected %.2f, got %.2f", tc.expected, got)
			}
		})
	}
}

func TestRefreshNotConfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Connection.Token = ""
	cfgStore := config.NewStore("unused", cfg)
	stateStore := state.New()

	refresher := New(nil, nil, cfgStore, nil, stateStore, slog.New(slog.NewTextHandler(io.Discard, nil)))
	updates := 0
	refresher.SetUpdateCallback(func() { updates++ })

	err := refresher.Refresh(context.Background())
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected not configured error, got %v", err)
	}
	if updates != 1 {
		t.Fatalf("expected update callback to be called once, got %d", updates)
	}
	snap := stateStore.Snapshot()
	if !snap.NotConfigured {
		t.Fatalf("expected state to be not configured")
	}
	if snap.LastError != "" {
		t.Fatalf("expected no error stored when not configured")
	}
}

func TestRefreshSuccess(t *testing.T) {
	client := newTestClient(t, http.StatusOK, `{"data":{"usage":12.34,"usage_daily":1.1,"usage_weekly":2.2,"usage_monthly":3.3,"id":"key-id"}}`)

	cfg := config.DefaultConfig()
	cfg.Connection.Token = "token"
	cfgStore := config.NewStore("unused", cfg)
	stateStore := state.New()
	cachePath := filepath.Join(t.TempDir(), cache.CacheFileName)
	cacheStore := cache.NewStore(cachePath)

	refresher := New(client, cacheStore, cfgStore, nil, stateStore, slog.New(slog.NewTextHandler(io.Discard, nil)))
	updates := 0
	refresher.SetUpdateCallback(func() { updates++ })

	before := time.Now().UTC()
	if err := refresher.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if updates != 1 {
		t.Fatalf("expected update callback to be called once, got %d", updates)
	}

	snap := stateStore.Snapshot()
	if snap.NotConfigured {
		t.Fatalf("expected configured state")
	}
	if snap.LastError != "" {
		t.Fatalf("expected no error after success")
	}
	if snap.Usage.Total != 12.34 {
		t.Fatalf("unexpected usage total: %v", snap.Usage.Total)
	}
	if snap.LastSuccessAt.IsZero() || snap.LastSuccessAt.Before(before) {
		t.Fatalf("expected last success time to be set")
	}

	loaded, err := cache.LoadFromPath(cachePath)
	if err != nil {
		t.Fatalf("cache load failed: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected cache to be saved")
	}
	if loaded.TotalUsage != 12.34 {
		t.Fatalf("expected cached total usage")
	}
	if loaded.KeyID != "key-id" {
		t.Fatalf("expected cached key id")
	}
	if loaded.KeyHash != util.TokenHash("token") {
		t.Fatalf("expected cached key hash")
	}
}

func TestRefreshUnauthorized(t *testing.T) {
	client := newTestClient(t, http.StatusUnauthorized, "unauthorized")

	cfg := config.DefaultConfig()
	cfg.Connection.Token = "token"
	cfgStore := config.NewStore("unused", cfg)
	stateStore := state.New()
	cachePath := filepath.Join(t.TempDir(), cache.CacheFileName)
	cacheStore := cache.NewStore(cachePath)

	refresher := New(client, cacheStore, cfgStore, nil, stateStore, slog.New(slog.NewTextHandler(io.Discard, nil)))
	updates := 0
	refresher.SetUpdateCallback(func() { updates++ })

	err := refresher.Refresh(context.Background())
	if !errors.Is(err, openrouter.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if updates != 1 {
		t.Fatalf("expected update callback to be called once, got %d", updates)
	}
	snap := stateStore.Snapshot()
	if snap.LastError == "" {
		t.Fatalf("expected error stored in state")
	}
	if _, err := os.Stat(cachePath); err == nil {
		t.Fatalf("expected cache not to be saved on error")
	}
}

func newTestClient(t *testing.T, status int, body string) *openrouter.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/key" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth == "" {
			t.Errorf("missing authorization header")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return openrouter.NewClient(server.URL, server.Client(), nil)
}
