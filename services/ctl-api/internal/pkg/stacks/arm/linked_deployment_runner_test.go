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

func TestGetDefaultRunnerDeployment_CustomDataUsesSelectedLocation(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()

	dep := tmpl.getDefaultRunnerDeployment(inp, nil, armScope{})
	vmss := dep["properties"].(map[string]any)["template"].(map[string]any)["resources"].([]any)[0].(map[string]any)
	got := vmss["properties"].(map[string]any)["virtualMachineProfile"].(map[string]any)["osProfile"].(map[string]any)["customData"]
	want := "[base64(replace(parameters('customData'), '__NUON_LOCATION__', parameters('location')))]"
	if got != want {
		t.Errorf("customData = %v, want %s", got, want)
	}

	script := dep["properties"].(map[string]any)["parameters"].(map[string]any)["customData"].(map[string]any)["value"].(string)
	if !strings.Contains(script, "AWS_REGION=__NUON_LOCATION__") {
		t.Error("cloud-init bakes a region instead of the location token")
	}
}

func TestGetDefaultRunnerDeployment_VMSizeIsRootParameter(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.Settings.AWSInstanceType = "t3.medium"

	dep := tmpl.getDefaultRunnerDeployment(inp, nil, armScope{})
	if got := runnerVMSSSKUName(t, dep); got != "[parameters('runnerVmSize')]" {
		t.Errorf("expected sku to reference runnerVmSize, got %q", got)
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	got := params["runnerVmSize"].(map[string]any)["value"]
	if got != "[parameters('runnerVmSize')]" {
		t.Errorf("expected runnerVmSize passthrough, got %v", got)
	}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate: %v", err)
	}
	p, ok := armTmpl.Parameters["runnerVmSize"]
	if !ok {
		t.Fatal("root template missing runnerVmSize")
	}
	if p.DefaultValue != app.DefaultAzureInstanceType {
		t.Errorf("default = %v, want %q", p.DefaultValue, app.DefaultAzureInstanceType)
	}

	sub, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("getAzureTemplate subscription: %v", err)
	}
	if _, ok := sub.Parameters["runnerVmSize"]; !ok {
		t.Fatal("subscription-scoped template missing runnerVmSize parameter")
	}
	if _, ok := sub.Variables["runnerVmSize"]; ok {
		t.Error("runnerVmSize must stay a parameter at subscription scope so the portal form exposes it")
	}
}

func TestGetAzureTemplate_VMSizeFromRunnerConfig(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.ConfiguredRunnerInstanceType = "Standard_D4s_v3"

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate: %v", err)
	}
	p := armTmpl.Parameters["runnerVmSize"]
	if p.DefaultValue != "Standard_D4s_v3" {
		t.Errorf("default = %v, want Standard_D4s_v3", p.DefaultValue)
	}
}

func TestGetRunnerLinkedDeployment_CustomTemplateReceivesVMSize(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()
	inp.ConfiguredRunnerInstanceType = "Standard_D4s_v3"
	inp.RunnerNestedStackTemplateURL = runnerTemplateServer(t, `{"parameters":{"runnerVmSize":{"type":"string"}},"resources":[]}`)

	dep, customerParams, err := tmpl.getRunnerLinkedDeployment(inp, nil, armScope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := dep["properties"].(map[string]any)["parameters"].(map[string]any)
	got := params["runnerVmSize"].(map[string]any)["value"]
	if got != "[parameters('runnerVmSize')]" {
		t.Errorf("expected runnerVmSize passthrough, got %v", got)
	}
	if customerParams["runnerVmSize"].DefaultValue != "Standard_D4s_v3" {
		t.Errorf("root default = %v, want Standard_D4s_v3", customerParams["runnerVmSize"].DefaultValue)
	}
}
