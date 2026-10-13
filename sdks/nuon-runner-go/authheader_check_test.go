package nuonrunner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestPublicEndpointsSendNoAuthHeader(t *testing.T) {
	var mu sync.Mutex
	authByPath := map[string]string{}
	queryByPath := map[string]string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		authByPath[r.URL.Path] = r.Header.Get("Authorization")
		queryByPath[r.URL.Path] = r.URL.RawQuery
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/runners/rnr_test/settings" {
			w.Write([]byte(`{}`))
			return
		}
		if r.URL.Path == "/v1/telemetry/access-token" || r.URL.Path == "/v1/installs/inl_test/telemetry/access-token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token":"access.jwt","token_type":"Bearer","expires_in":600}`))
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("[]"))
		default:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("{}"))
		}
	}))
	defer srv.Close()

	c, err := New(WithURL(srv.URL), WithAuthToken("secret-token"), WithRunnerID("rnr_test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	// authenticated call → must carry the bearer token
	if _, err := c.CreateHeartBeat(ctx, &models.ServiceCreateRunnerHeartBeatRequest{}); err != nil {
		t.Fatalf("CreateHeartBeat: %v", err)
	}
	if _, err := c.CreateTelemetryAccessToken(ctx, "https://relay.example.com/acme"); err != nil {
		t.Fatalf("CreateTelemetryAccessToken: %v", err)
	}
	if _, err := c.CreateInstallTelemetryAccessToken(ctx, "inl_test", "https://relay.example.com/acme"); err != nil {
		t.Fatalf("CreateInstallTelemetryAccessToken: %v", err)
	}
	if _, err := c.GetSettings(ctx); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	// public calls → must NOT carry an Authorization header
	_, _ = c.GetProcessShutdowns(ctx, "prc_test")
	_, _ = c.RunnerAuthAWS(ctx, &models.ServiceRunnerAuthAWSRequest{})

	mu.Lock()
	defer mu.Unlock()

	hbPath := "/v1/runners/rnr_test/heart-beats"
	if got := authByPath[hbPath]; got != "Bearer secret-token" {
		t.Errorf("authenticated heartbeat: got Authorization %q, want %q", got, "Bearer secret-token")
	}
	if got := authByPath["/v1/telemetry/access-token"]; got != "Bearer secret-token" {
		t.Errorf("telemetry access token: got Authorization %q, want %q", got, "Bearer secret-token")
	}
	if got := queryByPath["/v1/telemetry/access-token"]; got != "relay_endpoint=https%3A%2F%2Frelay.example.com%2Facme" {
		t.Errorf("token request was not bound to the selected relay: %q", got)
	}
	if got := authByPath["/v1/installs/inl_test/telemetry/access-token"]; got != "Bearer secret-token" {
		t.Errorf("install telemetry access token: got Authorization %q, want %q", got, "Bearer secret-token")
	}
	if got := queryByPath["/v1/installs/inl_test/telemetry/access-token"]; got != "relay_endpoint=https%3A%2F%2Frelay.example.com%2Facme" {
		t.Errorf("install token request was not bound to the selected relay: %q", got)
	}
	if got := queryByPath["/v1/runners/rnr_test/settings"]; got != "" {
		t.Errorf("settings request has unexpected query parameters: %q", got)
	}

	shutdownPath := "/v1/runners/rnr_test/processes/prc_test/shutdowns"
	if got := authByPath[shutdownPath]; got != "" {
		t.Errorf("public shutdowns: got Authorization %q, want empty", got)
	}

	authPath := "/v1/runner-auth/aws"
	if got := authByPath[authPath]; got != "" {
		t.Errorf("public runner-auth: got Authorization %q, want empty", got)
	}
}

func TestAuthTokenFileRotationAndFailure(t *testing.T) {
	var headers []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = append(headers, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.Write([]byte(`{"access_token":"relay.jwt","token_type":"Bearer","expires_in":600}`))
			return
		}
		w.Write([]byte(`{"enabled":true,"relay_endpoint":"https://relay.example.com","resource_attributes":{"nuon.install.name":"acme"}}`))
	}))
	defer srv.Close()

	directory := t.TempDir()
	path := filepath.Join(directory, "token")
	for name, token := range map[string]string{"before": "old-token\n", "after": "new-token\n"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(token), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("before", path); err != nil {
		t.Fatal(err)
	}
	c, err := New(WithURL(srv.URL), WithAuthToken("must-not-fall-back"), WithAuthTokenFile(path))
	if err != nil {
		t.Fatal(err)
	}
	settings, err := c.GetInstallTelemetryCollectorSettings(context.Background(), "inl_test")
	if err != nil || !settings.Enabled || settings.ResourceAttributes["nuon.install.name"] != "acme" {
		t.Fatalf("get collector settings: settings=%v error=%v", settings, err)
	}
	if err := os.Symlink("after", path+".next"); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path+".next", path); err != nil {
		t.Fatal(err)
	}
	token, err := c.CreateInstallTelemetryAccessToken(context.Background(), "inl_test", settings.RelayEndpoint)
	if err != nil || token.AccessToken != "relay.jwt" {
		t.Fatalf("issue token: response=%v error=%v", token, err)
	}
	if !slices.Equal(headers, []string{"Bearer old-token", "Bearer new-token"}) {
		t.Fatalf("projected credential rotation not picked up: %v", headers)
	}
	for _, invalid := range []string{"", "secret token", strings.Repeat("x", 16*1024+1)} {
		if err := os.WriteFile(filepath.Join(directory, "after"), []byte(invalid), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := c.GetInstallTelemetryCollectorSettings(context.Background(), "inl_test"); err == nil {
			t.Fatal("invalid file did not fail closed")
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetInstallTelemetryCollectorSettings(context.Background(), "inl_test"); err == nil {
		t.Fatal("missing file did not fail closed")
	}
	if len(headers) != 2 {
		t.Fatalf("invalid credential sent an authenticated request: %v", headers)
	}
}
