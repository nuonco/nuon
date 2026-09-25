package activities

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func TestRenderARMStackTemplate_UIDefWithoutWrapper(t *testing.T) {
	a := &Activities{cfg: &internal.Config{}}
	res, err := a.RenderARMStackTemplate(context.Background(), &RenderARMStackTemplateRequest{
		Input: stacks.TemplateInput{
			Install: &app.Install{
				ID:    "test-install-id-00000001",
				AppID: "test-app-id-000000000001",
				AzureAccount: &app.AzureAccount{
					Location: "eastus",
				},
			},
			CloudFormationStackVersion: &app.InstallStackVersion{
				PhoneHomeURL:            "https://api.example.com/phone-home/phid-test",
				QuickLinkUIDefBucketKey: "templates/inl-test/ist-test-ui.json",
			},
			Runner: &app.Runner{
				ID:    "test-runner-id-0000000001",
				OrgID: "test-org-id-00000000001",
			},
			Settings: &app.RunnerGroupSettings{
				ContainerImageURL: "example.com/runner",
				ContainerImageTag: "latest",
			},
			AppCfg:          &app.AppConfig{},
			DeploymentScope: app.StackDeploymentScopeSubscription,
		},
	})
	if err != nil {
		t.Fatalf("RenderARMStackTemplate: %v", err)
	}
	if len(res.QuickLinkWrapperJSON) != 0 {
		t.Fatal("rendered a wrapper without QuickLinkBucketKey")
	}
	if len(res.QuickLinkUIDefJSON) == 0 {
		t.Fatal("UI definition was not rendered")
	}

	var uiDef map[string]any
	if err := json.Unmarshal(res.QuickLinkUIDefJSON, &uiDef); err != nil {
		t.Fatalf("unmarshal UI definition: %v", err)
	}
	params := uiDef["parameters"].(map[string]any)
	steps := params["steps"].([]any)
	if len(steps) != 1 {
		t.Fatalf("steps = %d, want the runner step", len(steps))
	}
	elements := steps[0].(map[string]any)["elements"].([]any)
	element := elements[0].(map[string]any)
	if element["type"] != "Microsoft.Compute.SizeSelector" {
		t.Errorf("runner element type = %v", element["type"])
	}
	if got := params["outputs"].(map[string]any)["runnerVmSize"]; got != "[steps('runner').runnerVmSize]" {
		t.Errorf("runnerVmSize output = %v", got)
	}
}
