package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ConfigFileName)
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := DefaultConfig()
	if cfg.Updates.Period != def.Updates.Period || cfg.Updates.UpdateOnStart != def.Updates.UpdateOnStart {
		t.Fatalf("expected defaults, got %+v", cfg)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ConfigFileName)
	cfg := Config{
		Connection:    ConnectionConfig{Token: "token"},
		Updates:       UpdatesConfig{Period: "15m", UpdateOnStart: false},
		Notifications: NotificationsConfig{Enabled: true, OnUpdateSpent: false, OnError: true},
		Logging:       LoggingConfig{Level: "debug"},
	}
	if err := SaveToPath(path, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Connection.Token != cfg.Connection.Token {
		t.Fatalf("token mismatch: %v", loaded.Connection.Token)
	}
	if loaded.Updates.Period != cfg.Updates.Period {
		t.Fatalf("period mismatch: %v", loaded.Updates.Period)
	}
	if loaded.Notifications.Enabled != cfg.Notifications.Enabled {
		t.Fatalf("notifications mismatch")
	}
	if loaded.Logging.Level != cfg.Logging.Level {
		t.Fatalf("logging mismatch")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestNormalizeInvalidPeriod(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Updates.Period = "bogus"
	Normalize(&cfg)
	if cfg.Updates.Period == "bogus" {
		t.Fatalf("expected period to be normalized")
	}
}
