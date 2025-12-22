package summary

import (
	"strings"
	"testing"
	"time"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/openrouter"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/util"
)

func TestTooltipNotConfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	snap := state.Snapshot{}
	if got := Tooltip(cfg, snap); got != "Set token in Settings" {
		t.Fatalf("expected not configured message, got %q", got)
	}

	cfg.Connection.Token = "token"
	snap.NotConfigured = true
	if got := Tooltip(cfg, snap); got != "Set token in Settings" {
		t.Fatalf("expected not configured message, got %q", got)
	}
}

func TestTooltipContent(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Connection.Token = "token"
	when := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	daily := 1.23
	weekly := 2.34
	monthly := 3.45
	snap := state.Snapshot{
		LastSuccessAt: when,
		Usage: openrouter.Usage{
			Total:   10.5,
			Daily:   &daily,
			Weekly:  &weekly,
			Monthly: &monthly,
		},
		LastError: "boom",
	}

	lines := []string{
		"Daily: " + util.FormatUSD(daily),
		"Weekly: " + util.FormatUSD(weekly),
		"Monthly: " + util.FormatUSD(monthly),
		"Total: " + util.FormatUSD(10.5),
		"Updated: " + util.FormatTime(when),
		"ERROR: boom (stale)",
	}
	want := strings.Join(lines, "\n")

	if got := Tooltip(cfg, snap); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFormatUsageNil(t *testing.T) {
	if got := formatUsage(nil); got != "N/A" {
		t.Fatalf("expected N/A, got %q", got)
	}
}
