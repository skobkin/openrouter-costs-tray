package logging

import (
	"bytes"
	"strings"
	"testing"
)

type nopWriteCloser struct {
	*bytes.Buffer
}

func (n nopWriteCloser) Close() error {
	return nil
}

func TestOutputEnableWriterWrites(t *testing.T) {
	logger, _, output := NewLogger("info")
	buf := &bytes.Buffer{}
	if err := output.EnableWriter(nopWriteCloser{Buffer: buf}); err != nil {
		t.Fatalf("enable writer: %v", err)
	}
	logger.Info("hello", "foo", "bar")
	got := buf.String()
	if !strings.Contains(got, "\"msg\":\"hello\"") {
		t.Fatalf("expected log message in output, got %q", got)
	}
	if !strings.Contains(got, "\"foo\":\"bar\"") {
		t.Fatalf("expected log attributes in output, got %q", got)
	}
}

func TestOutputDisableFileStopsWriting(t *testing.T) {
	logger, _, output := NewLogger("info")
	buf := &bytes.Buffer{}
	if err := output.EnableWriter(nopWriteCloser{Buffer: buf}); err != nil {
		t.Fatalf("enable writer: %v", err)
	}
	logger.Info("first")
	before := buf.String()
	if err := output.DisableFile(); err != nil {
		t.Fatalf("disable file: %v", err)
	}
	logger.Info("second")
	if buf.String() != before {
		t.Fatalf("expected no new output after disable, got %q", buf.String())
	}
}
