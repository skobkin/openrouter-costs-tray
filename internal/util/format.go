package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func FormatUSD(value float64) string {
	if value < 0 {
		value = 0
	}
	switch {
	case value < 1:
		return fmt.Sprintf("$%.4f", value)
	case value < 10:
		return fmt.Sprintf("$%.3f", value)
	default:
		return fmt.Sprintf("$%.2f", value)
	}
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Local().Format("2006-01-02 15:04")
}

func TokenHash(token string) string {
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
