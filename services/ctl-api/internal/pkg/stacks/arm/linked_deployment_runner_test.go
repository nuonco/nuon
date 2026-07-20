package arm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func runnerTemplateServer(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/runner.json"
}

func TestGetRunnerLinkedDeployment_CustomTemplateMissingIdentityParam(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, `{"parameters":{"nuonInstallID":{"type":"string"}},"resources":[]}`)

	ids := []azureOperationIdentity{{roleName: "prov", suffix: "provision", kind: "provision"}}
	_, _, err := tmpl.getRunnerLinkedDeployment(inp, ids)
	if err == nil {
		t.Fatal("expected error for custom template without userAssignedIdentities param")
	}
}

func TestGetRunnerLinkedDeployment_CustomTemplateAttachesIdentities(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, `{"parameters":{"nuonInstallID":{"type":"string"},"userAssignedIdentities":{"type":"object"}},"resources":[]}`)

	ids := []azureOperationIdentity{{roleName: "prov", suffix: "provision", kind: "provision"}}
	dep, _, err := tmpl.getRunnerLinkedDeployment(inp, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	if _, ok := params["userAssignedIdentities"]; !ok {
		t.Error("expected userAssignedIdentities injected into deployment params")
	}
	deps := dep["dependsOn"].([]string)
	if len(deps) < 2 {
		t.Errorf("expected identity dependsOn appended, got %v", deps)
	}
}
