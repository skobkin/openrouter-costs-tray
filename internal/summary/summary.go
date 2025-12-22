package summary

import (
	"strings"

	"openrouter-costs-tray/internal/config"
	"openrouter-costs-tray/internal/state"
	"openrouter-costs-tray/internal/util"
)

func Tooltip(cfg config.Config, snap state.Snapshot) string {
	if cfg.Connection.Token == "" || snap.NotConfigured {
		return "Set token in Settings"
	}
	lines := []string{
		"Daily: " + formatUsage(snap.Usage.Daily),
		"Weekly: " + formatUsage(snap.Usage.Weekly),
		"Monthly: " + formatUsage(snap.Usage.Monthly),
		"Total: " + util.FormatUSD(snap.Usage.Total),
		"Updated: " + util.FormatTime(snap.LastSuccessAt),
	}
	if snap.LastError != "" {
		lines = append(lines, "ERROR: "+snap.LastError+" (stale)")
	}
	return strings.Join(lines, "\n")
}

func formatUsage(value *float64) string {
	if value == nil {
		return "N/A"
	}
	return util.FormatUSD(*value)
}
