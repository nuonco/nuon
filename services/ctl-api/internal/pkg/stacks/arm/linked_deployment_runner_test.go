package arm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	_, _, err := tmpl.getRunnerLinkedDeployment(inp, ids, armScope{})
	if err == nil {
		t.Fatal("expected error for custom template without userAssignedIdentities param")
	}
}

func TestGetRunnerLinkedDeployment_CustomTemplateAttachesIdentities(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, `{"parameters":{"nuonInstallID":{"type":"string"},"userAssignedIdentities":{"type":"object"}},"resources":[]}`)

	ids := []azureOperationIdentity{{roleName: "prov", suffix: "provision", kind: "provision"}}
	dep, _, err := tmpl.getRunnerLinkedDeployment(inp, ids, armScope{})
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

func runnerVMSSSKUName(t *testing.T, dep map[string]any) string {
	t.Helper()
	tmpl := dep["properties"].(map[string]any)["template"].(map[string]any)
	vmss := tmpl["resources"].([]any)[0].(map[string]any)
	return vmss["sku"].(map[string]any)["name"].(string)
}

func TestGetDefaultRunnerDeployment_VMSizeDefaultsToPlatformDefault(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.Settings.AWSInstanceType = "t3.medium"

	dep := tmpl.getDefaultRunnerDeployment(inp, nil, armScope{})
	if got := runnerVMSSSKUName(t, dep); got != app.DefaultAzureInstanceType {
		t.Errorf("expected sku %q, got %q", app.DefaultAzureInstanceType, got)
	}
}

func TestGetDefaultRunnerDeployment_VMSizeFromRunnerConfig(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.ConfiguredRunnerInstanceType = "Standard_D4s_v3"

	dep := tmpl.getDefaultRunnerDeployment(inp, nil, armScope{})
	if got := runnerVMSSSKUName(t, dep); got != "Standard_D4s_v3" {
		t.Errorf("expected sku Standard_D4s_v3, got %q", got)
	}
}

func TestGetRunnerLinkedDeployment_CustomTemplateReceivesVMSize(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.ConfiguredRunnerInstanceType = "Standard_D4s_v3"
	inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, `{"parameters":{"runnerVmSize":{"type":"string"}},"resources":[]}`)

	dep, _, err := tmpl.getRunnerLinkedDeployment(inp, nil, armScope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	got := params["runnerVmSize"].(map[string]any)["value"]
	if got != "Standard_D4s_v3" {
		t.Errorf("expected runnerVmSize Standard_D4s_v3, got %v", got)
	}
}
