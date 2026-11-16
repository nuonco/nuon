package nuon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCLIConfigHandlesNonJSONErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/general/cli-config" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("<!doctype html><title>403</title>403 Forbidden"))
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetCLIConfig(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "Failed to retrieve cli-config. Request return 403" {
		t.Fatalf("unexpected error: %q", err)
	}
	if strings.Contains(err.Error(), "TextConsumer") {
		t.Fatalf("unexpected decoder error: %q", err)
	}
	if !IsForbidden(err) {
		t.Fatalf("expected forbidden error: %v", err)
	}
}

func TestGetCLIConfigPreservesJSONErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"description":"CLI configuration is temporarily unavailable","error":"unavailable"}`))
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetCLIConfig(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "Failed to retrieve cli-config. Request return 503" {
		t.Fatalf("unexpected error: %q", err)
	}
	if !IsServerError(err) {
		t.Fatalf("expected server error: %v", err)
	}
}
