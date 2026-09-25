package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestReprovisionInstallSandboxSkipComponents(t *testing.T) {
	requests := make(chan models.ServiceReprovisionInstallSandboxRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/installs/install-id/reprovision-sandbox" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req models.ServiceReprovisionInstallSandboxRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		requests <- req

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"workflow_id":"workflow-id"}`)
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL), WithOrgID("org-id"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ReprovisionInstallSandbox(context.Background(), "install-id"); err != nil {
		t.Fatal(err)
	}
	if req := <-requests; req.SkipComponents {
		t.Fatal("skip_components = true without option")
	}

	if _, err := client.ReprovisionInstallSandbox(context.Background(), "install-id", true); err != nil {
		t.Fatal(err)
	}
	if req := <-requests; !req.SkipComponents {
		t.Fatal("skip_components = false with option enabled")
	}
}
