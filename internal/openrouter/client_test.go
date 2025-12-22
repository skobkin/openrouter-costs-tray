package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchUsageSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"key123","usage":12.34,"usage_daily":1.2,"usage_weekly":5.6,"usage_monthly":7.8}}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client(), nil)
	usage, err := client.FetchUsage(context.Background(), "token")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if usage.Total != 12.34 {
		t.Fatalf("total mismatch: %v", usage.Total)
	}
	if usage.Daily == nil || *usage.Daily != 1.2 {
		t.Fatalf("daily mismatch")
	}
	if usage.KeyID != "key123" {
		t.Fatalf("key id mismatch")
	}
}

func TestFetchUsageUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client(), nil)
	_, err := client.FetchUsage(context.Background(), "token")
	if err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestFetchUsageMalformed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client(), nil)
	_, err := client.FetchUsage(context.Background(), "token")
	if err == nil {
		t.Fatalf("expected error")
	}
}
