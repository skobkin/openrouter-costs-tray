package logging

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Output struct {
	writer *multiWriter
}

type WriteCloser interface {
	io.Writer
	io.Closer
}

func NewLogger(level string) (*slog.Logger, *slog.LevelVar, *Output) {
	levelVar := &slog.LevelVar{}
	SetLevel(levelVar, level)
	writer := &multiWriter{stdout: os.Stdout}
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: levelVar})
	return slog.New(handler), levelVar, &Output{writer: writer}
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

func (o *Output) EnableFile(path string) error {
	if o == nil || o.writer == nil {
		return errors.New("log output not initialized")
	}
	if strings.TrimSpace(path) == "" {
		return errors.New("log file path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	//nolint:gosec // path comes from app config directory, not user input
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	o.writer.setFile(file)
	return nil
}

func (o *Output) EnableWriter(writer WriteCloser) error {
	if o == nil || o.writer == nil {
		return errors.New("log output not initialized")
	}
	if writer == nil {
		return errors.New("log output writer is nil")
	}
	o.writer.setFile(writer)
	return nil
}

func (o *Output) DisableFile() error {
	if o == nil || o.writer == nil {
		return nil
	}
	return o.writer.clearFile()
}

type multiWriter struct {
	mu     sync.Mutex
	stdout io.Writer
	file   WriteCloser
}

func (w *multiWriter) Write(p []byte) (int, error) {
	if w == nil || w.stdout == nil {
		return 0, errors.New("log output not initialized")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.stdout.Write(p)
	if w.file != nil {
		if _, fileErr := w.file.Write(p); fileErr != nil && err == nil {
			err = fileErr
		}
	}
	return n, err
}

func (w *multiWriter) setFile(file WriteCloser) {
	w.mu.Lock()
	old := w.file
	w.file = file
	w.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
}

func (w *multiWriter) clearFile() error {
	w.mu.Lock()
	old := w.file
	w.file = nil
	w.mu.Unlock()
	if old != nil {
		return old.Close()
	}
	return nil
}
