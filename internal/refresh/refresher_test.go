package refresh

import (
	"testing"

	"openrouter-costs-tray/internal/cache"
)

func TestComputeDelta(t *testing.T) {
	prev := &cache.CostsCache{TotalUsage: 10, KeyHash: "abc"}

	cases := []struct {
		name     string
		prev     *cache.CostsCache
		hash     string
		current  float64
		expected float64
	}{
		{"no prev", nil, "abc", 10, 0},
		{"hash mismatch", prev, "def", 12, 0},
		{"increase", prev, "abc", 12, 2},
		{"equal", prev, "abc", 10, 0},
		{"decrease", prev, "abc", 8, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeDelta(tc.prev, tc.hash, tc.current)
			if got != tc.expected {
				t.Fatalf("expected %.2f, got %.2f", tc.expected, got)
			}
		})
	}
}
