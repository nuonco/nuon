package arm

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	dep, params, err := tmpl.getRunnerLinkedDeployment(inp, nil, armScope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := runnerVMSSSKUName(t, dep); got != "[parameters('runnerVmSize')]" {
		t.Errorf("sku name = %q", got)
	}
	if got := params["runnerVmSize"].DefaultValue; got != app.DefaultAzureInstanceType {
		t.Errorf("runnerVmSize default = %v, want %q", got, app.DefaultAzureInstanceType)
	}
}

func TestGetDefaultRunnerDeployment_VMSizeFromRunnerConfig(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.ConfiguredRunnerInstanceType = "Standard_D4s_v3"

	_, params, err := tmpl.getRunnerLinkedDeployment(inp, nil, armScope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := params["runnerVmSize"].DefaultValue; got != "Standard_D4s_v3" {
		t.Errorf("runnerVmSize default = %v, want Standard_D4s_v3", got)
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
	if got != "[parameters('runnerVmSize')]" {
		t.Errorf("runnerVmSize = %v, want the hoisted parameter", got)
	}
}

func TestBuildRunnerCustomData_AuthResource(t *testing.T) {
	inp := minimalTemplateInput()

	without := (&Templates{cfg: &internal.Config{}}).buildRunnerCustomData(inp)
	if strings.Contains(without, "AZURE_RUNNER_AUTH_RESOURCE") {
		t.Fatal("unset audience wrote a resource line")
	}

	const audience = "api://11111111-1111-1111-1111-111111111111"
	with := (&Templates{cfg: &internal.Config{AzureRunnerAuthAudience: audience}}).buildRunnerCustomData(inp)
	if !strings.Contains(with, "AZURE_RUNNER_AUTH_RESOURCE="+audience+"\n") {
		t.Fatalf("custom data missing the audience:\n%s", with)
	}

	invalid := (&Templates{cfg: &internal.Config{AzureRunnerAuthAudience: "https://management.azure.com/"}}).buildRunnerCustomData(inp)
	if strings.Contains(invalid, "AZURE_RUNNER_AUTH_RESOURCE") {
		t.Fatal("ARM audience was written into the runner env")
	}
}
