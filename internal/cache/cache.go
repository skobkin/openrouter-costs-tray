package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const CacheFileName = "costs_cache.json"

const SchemaVersion = "1"

type CostsCache struct {
	SchemaVersion string    `json:"schema_version,omitempty"`
	LastSuccessAt time.Time `json:"last_success_at"`
	TotalUsage    float64   `json:"total_usage"`
	DailyUsage    *float64  `json:"daily_usage,omitempty"`
	WeeklyUsage   *float64  `json:"weekly_usage,omitempty"`
	MonthlyUsage  *float64  `json:"monthly_usage,omitempty"`
	KeyHash       string    `json:"key_hash,omitempty"`
	KeyID         string    `json:"key_id,omitempty"`
}

func DefaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "openrouter-cost-tray"), nil
}

func DefaultCachePath() (string, error) {
	dir, err := DefaultCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CacheFileName), nil
}

func LoadFromPath(path string) (*CostsCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var cache CostsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func SaveToPath(path string, cache CostsCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o600)
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (*CostsCache, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return LoadFromPath(s.path)
}

func (s *Store) Save(cache CostsCache) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return SaveToPath(s.path, cache)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-cache-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(path)
		return os.Rename(tmpName, path)
	}
	return nil
}
