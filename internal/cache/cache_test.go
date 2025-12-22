package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingReturnsNil(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, CacheFileName)
	cached, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cached != nil {
		t.Fatalf("expected nil cache, got %+v", cached)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, CacheFileName)
	now := time.Now().UTC()
	val := 1.23
	cache := CostsCache{
		SchemaVersion: SchemaVersion,
		LastSuccessAt: now,
		TotalUsage:    10.5,
		DailyUsage:    &val,
		KeyHash:       "hash",
		KeyID:         "key",
	}
	if err := SaveToPath(path, cache); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected cache")
	}
	if loaded.TotalUsage != cache.TotalUsage {
		t.Fatalf("total mismatch")
	}
	if loaded.KeyHash != cache.KeyHash {
		t.Fatalf("key hash mismatch")
	}
	if loaded.DailyUsage == nil || *loaded.DailyUsage != *cache.DailyUsage {
		t.Fatalf("daily usage mismatch")
	}
}
