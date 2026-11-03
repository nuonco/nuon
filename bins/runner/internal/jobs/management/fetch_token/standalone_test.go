package fetchtoken

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type gcpAuthClient struct {
	nuonrunner.Client
	req *models.ServiceRunnerAuthGCPRequest
}

func (c *gcpAuthClient) RunnerAuthGCP(_ context.Context, req *models.ServiceRunnerAuthGCPRequest) (*models.ServiceRunnerAuthGCPResponse, error) {
	c.req = req
	return &models.ServiceRunnerAuthGCPResponse{Authenticated: true, RunnerID: "runner123", Token: "nuon-token"}, nil
}

func TestFetchTokenGCP_SendsIdentityTokenAndRunnerIDOnly(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		w.Header().Set("Metadata-Flavor", "Google")
		switch {
		case strings.HasSuffix(r.URL.Path, "/service-accounts/default/identity"):
			w.Write([]byte("identity-jwt"))
		case strings.HasSuffix(r.URL.Path, "/service-accounts/default/token"):
			w.Write([]byte(`{"access_token":"access-token","expires_in":3600,"token_type":"Bearer"}`))
		case strings.HasSuffix(r.URL.Path, "/attributes/nuon_runner_id"):
			w.Write([]byte("runner123"))
		default:
			w.Write([]byte("value"))
		}
	}))
	defer srv.Close()

	t.Setenv("GCE_METADATA_HOST", srv.Listener.Addr().String())
	t.Setenv("CLOUD_PROVIDER", "gcp")
	t.Setenv("RUNNER_API_URL", "https://runner.example.com")

	client := &gcpAuthClient{}
	result, err := FetchToken(t.Context(), client, "", "")
	if err != nil {
		t.Fatalf("FetchToken() error = %v", err)
	}
	if result.Token != "nuon-token" {
		t.Fatalf("token = %q, want nuon-token", result.Token)
	}

	body, err := json.Marshal(client.req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if _, ok := sent["metadata"]; ok {
		t.Fatalf("request carried a metadata block with an access token: %v", sent["metadata"])
	}
	if sent["identity_token"] != "identity-jwt" {
		t.Fatalf("identity_token = %v, want identity-jwt", sent["identity_token"])
	}
	if sent["runner_id"] != "runner123" {
		t.Fatalf("runner_id = %v, want runner123", sent["runner_id"])
	}

	for _, p := range paths {
		if strings.HasSuffix(p, "/service-accounts/default/token") {
			t.Fatalf("runner requested an access token from the metadata server (%s)", p)
		}
	}
}
