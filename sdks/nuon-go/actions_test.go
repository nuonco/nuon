package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestCreateInstallActionWorkflowRunWithResponseReturnsWorkflowID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/installs/install-1/action-workflows/runs" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"workflow_id":"wkf-1"}`)
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.CreateInstallActionWorkflowRunWithResponse(context.Background(), "install-1", &models.ServiceCreateInstallActionWorkflowRunRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.WorkflowID != "wkf-1" {
		t.Fatalf("expected workflow ID wkf-1, got %q", resp.WorkflowID)
	}

	if err := client.CreateInstallActionWorkflowRun(context.Background(), "install-1", &models.ServiceCreateInstallActionWorkflowRunRequest{}); err != nil {
		t.Fatal(err)
	}
}
