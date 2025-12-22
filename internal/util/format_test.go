package util

import (
	"testing"
	"time"
)

func TestFormatUSD(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  string
	}{
		{"negative", -1, "$0.0000"},
		{"sub1", 0.5, "$0.5000"},
		{"sub10", 1.234, "$1.234"},
		{"tenPlus", 12.3, "$12.30"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatUSD(tc.value); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	if got := FormatTime(time.Time{}); got != "never" {
		t.Fatalf("expected never, got %q", got)
	}

	stamp := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	want := stamp.Local().Format("2006-01-02 15:04")
	if got := FormatTime(stamp); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTokenHash(t *testing.T) {
	if got := TokenHash(""); got != "" {
		t.Fatalf("expected empty hash, got %q", got)
	}

	const token = "abc"
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := TokenHash(token); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
