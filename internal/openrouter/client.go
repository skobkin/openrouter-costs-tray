package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://openrouter.ai/api/v1"

var ErrUnauthorized = errors.New("openrouter unauthorized")

// Usage represents the usage totals returned by the API.
type Usage struct {
	Total   float64
	Daily   *float64
	Weekly  *float64
	Monthly *float64
	KeyID   string
	Label   string
}

type Client struct {
	baseURL string
	http    *http.Client
	logger  *slog.Logger
}

func NewClient(baseURL string, httpClient *http.Client, logger *slog.Logger) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient, logger: logger}
}

func (c *Client) FetchUsage(ctx context.Context, token string) (Usage, error) {
	if strings.TrimSpace(token) == "" {
		return Usage{}, errors.New("token is empty")
	}
	url := c.baseURL + "/auth/key"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Usage{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return Usage{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Usage{}, err
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return Usage{}, ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Usage{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	usage, err := parseUsage(body)
	if err != nil {
		return Usage{}, err
	}
	return usage, nil
}

func parseUsage(body []byte) (Usage, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	var payload any
	if err := dec.Decode(&payload); err != nil {
		return Usage{}, err
	}
	var usageMap map[string]any
	if top, ok := payload.(map[string]any); ok {
		if data, ok := top["data"].(map[string]any); ok {
			if _, has := data["usage"]; has {
				usageMap = data
			}
		}
		if usageMap == nil {
			if key, ok := top["key"].(map[string]any); ok {
				if _, has := key["usage"]; has {
					usageMap = key
				}
			}
		}
	}
	if usageMap == nil {
		var ok bool
		usageMap, ok = findUsageMap(payload)
		if !ok {
			return Usage{}, errors.New("usage field not found in response")
		}
	}
	usage := Usage{}
	if total, ok := toFloat(usageMap["usage"]); ok {
		usage.Total = total
	} else {
		return Usage{}, errors.New("usage value missing or invalid")
	}
	if daily, ok := toFloat(usageMap["usage_daily"]); ok {
		usage.Daily = &daily
	}
	if weekly, ok := toFloat(usageMap["usage_weekly"]); ok {
		usage.Weekly = &weekly
	}
	if monthly, ok := toFloat(usageMap["usage_monthly"]); ok {
		usage.Monthly = &monthly
	}
	usage.KeyID = firstString(usageMap, "id", "key_id", "api_key_id")
	usage.Label = firstString(usageMap, "name", "label")
	return usage, nil
}

func findUsageMap(value any) (map[string]any, bool) {
	switch v := value.(type) {
	case map[string]any:
		if _, ok := v["usage"]; ok {
			return v, true
		}
		for _, child := range v {
			if found, ok := findUsageMap(child); ok {
				return found, true
			}
		}
	case []any:
		for _, child := range v {
			if found, ok := findUsageMap(child); ok {
				return found, true
			}
		}
	}
	return nil, false
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case nil:
		return 0, false
	case float64:
		return v, true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := json.Number(v).Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			case json.Number:
				return v.String()
			}
		}
	}
	return ""
}
