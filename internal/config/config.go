package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	AppDirName     = "openrouter-cost-tray"
	ConfigFileName = "config.json"
	LogFileName    = "openrouter-costs-tray.log"
)

var PeriodOptions = []string{"5m", "15m", "30m", "1h", "3h", "6h", "12h"}

type ConnectionConfig struct {
	Token string `json:"token"`
}

type UpdatesConfig struct {
	Period        string `json:"period"`
	UpdateOnStart bool   `json:"update_on_start"`
}

type NotificationsConfig struct {
	Enabled        bool `json:"enabled"`
	OnUpdateSpent  bool `json:"on_update_spent"`
	OnError        bool `json:"on_error"`
	OnStartSummary bool `json:"on_start_summary"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	ToFile bool   `json:"to_file"`
}

type Config struct {
	Connection    ConnectionConfig    `json:"connection"`
	Updates       UpdatesConfig       `json:"updates"`
	Notifications NotificationsConfig `json:"notifications"`
	Logging       LoggingConfig       `json:"logging"`
}

func DefaultConfig() Config {
	return Config{
		Connection: ConnectionConfig{
			Token: "",
		},
		Updates: UpdatesConfig{
			Period:        "30m",
			UpdateOnStart: true,
		},
		Notifications: NotificationsConfig{
			Enabled:        false,
			OnUpdateSpent:  true,
			OnError:        true,
			OnStartSummary: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			ToFile: false,
		},
	}
}

func Normalize(cfg *Config) {
	if cfg == nil {
		return
	}
	if !IsValidPeriod(cfg.Updates.Period) {
		cfg.Updates.Period = DefaultConfig().Updates.Period
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = DefaultConfig().Logging.Level
	}
}

func DefaultConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppDirName), nil
}

func DefaultConfigPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func LoadFromPath(path string) (Config, error) {
	//nolint:gosec // path comes from config/store, not user input
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := DefaultConfig()
			Normalize(&cfg)
			return cfg, nil
		}
		return DefaultConfig(), err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	Normalize(&cfg)
	return cfg, nil
}

func SaveToPath(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o600)
}

type Store struct {
	mu   sync.RWMutex
	cfg  Config
	path string
}

func NewStore(path string, cfg Config) *Store {
	return &Store{cfg: cfg, path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Store) Set(cfg Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

func (s *Store) Save() error {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()
	return SaveToPath(s.path, cfg)
}

func ParsePeriod(period string) (time.Duration, bool) {
	switch period {
	case "5m":
		return 5 * time.Minute, true
	case "15m":
		return 15 * time.Minute, true
	case "30m":
		return 30 * time.Minute, true
	case "1h":
		return time.Hour, true
	case "3h":
		return 3 * time.Hour, true
	case "6h":
		return 6 * time.Hour, true
	case "12h":
		return 12 * time.Hour, true
	default:
		return 0, false
	}
}

func IsValidPeriod(period string) bool {
	for _, p := range PeriodOptions {
		if p == period {
			return true
		}
	}
	return false
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-config-*")
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
