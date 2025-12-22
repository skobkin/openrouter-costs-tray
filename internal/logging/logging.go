package logging

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(level string) (*slog.Logger, *slog.LevelVar) {
	levelVar := &slog.LevelVar{}
	SetLevel(levelVar, level)
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: levelVar})
	return slog.New(handler), levelVar
}

func SetLevel(levelVar *slog.LevelVar, level string) {
	if levelVar == nil {
		return
	}
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		levelVar.Set(slog.LevelDebug)
	case "warn", "warning":
		levelVar.Set(slog.LevelWarn)
	case "error":
		levelVar.Set(slog.LevelError)
	default:
		levelVar.Set(slog.LevelInfo)
	}
}
