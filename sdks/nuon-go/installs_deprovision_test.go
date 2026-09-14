package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestRunnerDependentWorkflowErrorsIncludeResponseDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/installs/install-1/deprovision",
			"/v1/installs/install-1/deprovision-sandbox",
			"/v1/installs/install-1/reprovision-sandbox",
			"/v1/installs/install-1/inputs",
			"/v1/installs/install-1/components/deploy-all",
			"/v1/installs/install-1/components/teardown-all",
			"/v1/installs/install-1/components/component-1/teardown",
			"/v1/installs/install-1/runbooks/runbook-1/runs",
			"/v1/installs/install-1/action-workflows/runs":
		default:
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"error":"runner unavailable","user_error":true,"description":"Wait for the runner and try again."}`)
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	for name, deprovision := range map[string]func() error{
		"install": func() error {
			_, err := client.DeprovisionInstall(context.Background(), "install-1")
			return err
		},
		"sandbox": func() error {
			_, err := client.DeprovisionInstallSandbox(context.Background(), "install-1")
			return err
		},
		"reprovision sandbox": func() error {
			_, err := client.ReprovisionInstallSandbox(context.Background(), "install-1")
			return err
		},
		"deploy components": func() error {
			_, err := client.DeployInstallComponents(context.Background(), "install-1", "", false)
			return err
		},
		"teardown components": func() error {
			_, err := client.TeardownInstallComponents(context.Background(), "install-1")
			return err
		},
		"teardown component": func() error {
			_, err := client.TeardownInstallComponent(context.Background(), "install-1", "component-1", "")
			return err
		},
		"update inputs": func() error {
			_, err := client.UpdateInstallInputs(context.Background(), "install-1", &models.ServiceUpdateInstallInputsRequest{})
			return err
		},
		"runbook run": func() error {
			_, err := client.CreateInstallRunbookRun(context.Background(), "install-1", "runbook-1")
			return err
		},
		"action workflow run": func() error {
			return client.CreateInstallActionWorkflowRun(context.Background(), "install-1", &models.ServiceCreateInstallActionWorkflowRunRequest{})
		},
	} {
		t.Run(name, func(t *testing.T) {
			userErr, ok := ToUserError(deprovision())
			if !ok {
				t.Fatal("expected a user error")
			}
			if userErr.Description != "Wait for the runner and try again." {
				t.Fatalf("unexpected description: %q", userErr.Description)
			}
		})
	}
}

func TestDeprovisionDecodesWorkflowResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"workflow_id":"workflow-1"}`)
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.DeprovisionInstall(context.Background(), "install-1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.WorkflowID != "workflow-1" {
		t.Fatalf("unexpected workflow ID: %q", resp.WorkflowID)
	}
}
