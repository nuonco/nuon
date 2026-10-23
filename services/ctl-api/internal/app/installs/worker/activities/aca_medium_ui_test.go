package activities

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/arm"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const acaMediumDir = "/Users/fd/nuon/example-app-configs/aca-medium"

// TestAcaMediumUIDefinition runs the aca-medium stack through the same steps a
// sync and a subscription-scoped quick link use: parse the ui.json, upload it,
// merge it into the rendered createUiDefinition, and point the portal link at
// that rendered file.
func TestAcaMediumUIDefinition(t *testing.T) {
	if _, err := os.Stat(acaMediumDir); err != nil {
		t.Skip(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:      acaMediumDir,
		V:            validator.New(),
		SkipBranches: true,
		SkipInstalls: true,
		FileProcessor: func(name string, obj map[string]any) map[string]any {
			return obj
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cfg.Stack)
	require.Equal(t, "azure-bicep", cfg.Stack.Type)
	require.Equal(t, "subscription", cfg.Stack.DeploymentScope)
	require.Equal(t, "./arm/ui.json", cfg.Stack.UIDefinition)
	require.NotEmpty(t, cfg.Stack.UIDefinitionContents)
	require.NoError(t, config.ValidateAzureUIDefinition(cfg.Stack.Type, cfg.Stack.UIDefinition))
	require.NoError(t, config.ValidateCreateUIDefinition(cfg.Stack.UIDefinitionContents))

	stackCfg, err := build.StackConfig(cfg.Stack, "app-aca-medium", "cfg-aca-medium")
	require.NoError(t, err)
	require.NotNil(t, stackCfg.UIDefinition)
	require.Equal(t, config.CustomNestedStackStatusPending, stackCfg.UIDefinition.Status)
	original := stackCfg.UIDefinition.Contents

	uploader := &memoryBlobUploader{}
	const baseURL = "https://templates.example.com"
	const orgID = "org-aca"
	const appID = "app-aca-medium"
	stackCfg.OrgID = orgID
	stackCfg.AppID = appID
	require.NoError(t, cloudformation.UploadUIDefinition(ctx, uploader, baseURL, orgID, appID, stackCfg.UIDefinition))

	ui := stackCfg.UIDefinition
	require.Equal(t, config.CustomNestedStackStatusReady, ui.Status)
	require.Empty(t, ui.Contents)
	require.NotEmpty(t, ui.ContentsHash)
	wantKey := cloudformation.UIDefinitionS3Key(orgID, appID, ui.ContentsHash)
	require.Equal(t, []byte(original), uploader.blobs[wantKey])
	require.Equal(t, baseURL+"/"+wantKey, ui.SourceURL)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+wantKey {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(uploader.blobs[wantKey])
	}))
	t.Cleanup(server.Close)
	ui.SourceURL = server.URL + "/" + wantKey

	inp := acaMediumTemplateInput(cfg, ui)
	tmpl := arm.NewTemplates(arm.Params{Cfg: &internal.Config{}})

	rendered, checksum, err := tmpl.QuickLinkUIDefinition(inp)
	require.NoError(t, err)
	require.NotEmpty(t, checksum)
	require.NotEqual(t, original, string(rendered), "rendered UI should be the merged portal form, not the vendor file")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(rendered, &doc))
	params := doc["parameters"].(map[string]any)
	basics := indexUI(params["basics"].([]any))
	steps := indexSteps(params["steps"].([]any))
	outputs := params["outputs"].(map[string]any)

	require.Contains(t, steps, "networking")
	require.Equal(t, "10.100.0.0/22", steps["networking"]["addressSpace"]["defaultValue"])
	require.Equal(t, "Microsoft.Common.DropDown", steps["networking"]["clusterIngressEndpointType"]["type"])
	require.Equal(t, "Microsoft.Common.CheckBox", steps["networking"]["enableAzureFirewall"]["type"])
	require.Equal(t, "Microsoft.Common.CheckBox", steps["networking"]["enableFrontDoor"]["type"])
	require.Contains(t, steps, "compute")
	require.Equal(t, "Microsoft.Common.DropDown", steps["compute"]["availabilityZones"]["type"])

	require.Equal(t, "[steps('networking').addressSpace]", outputs["addressSpace"])
	require.Equal(t, "[steps('networking').clusterIngressEndpointType]", outputs["clusterIngressEndpointType"])
	require.Equal(t, "[steps('compute').availabilityZones]", outputs["availabilityZones"])
	require.Equal(t, "[and(equals(steps('networking').clusterIngressEndpointType, 'private'), steps('networking').enableAzureFirewall)]", outputs["enableAzureFirewall"])
	require.Equal(t, "[and(equals(steps('networking').clusterIngressEndpointType, 'public'), steps('networking').enableFrontDoor)]", outputs["enableFrontDoor"])

	configBasics := params["config"].(map[string]any)["basics"].(map[string]any)
	location := configBasics["location"].(map[string]any)
	require.Equal(t, []any{"eastus"}, location["allowedValues"])
	require.NotContains(t, basics, "location")

	armTmpl, _, err := tmpl.Template(inp)
	require.NoError(t, err)
	var root map[string]any
	require.NoError(t, json.Unmarshal(armTmpl, &root))
	rootParams := root["parameters"].(map[string]any)
	for _, name := range []string{
		"addressSpace",
		"clusterIngressEndpointType", "enableAzureFirewall", "enableFrontDoor", "availabilityZones",
	} {
		require.Contains(t, rootParams, name, "subscription stack is missing hoisted parameter %s", name)
	}

	stackVersion := inp.CloudFormationStackVersion
	stackVersion.TemplateURL = baseURL + "/" + stackVersion.AWSBucketKey
	wrapper, _, err := tmpl.QuickLinkWrapper(inp, stackVersion.TemplateURL)
	require.NoError(t, err)
	var wrapperDoc map[string]any
	require.NoError(t, json.Unmarshal(wrapper, &wrapperDoc))
	wrapperParams := wrapperDoc["parameters"].(map[string]any)
	for name := range rootParams {
		require.Contains(t, wrapperParams, name)
	}
	for _, name := range []string{"resourceGroupName", "virtualNetworkName"} {
		t.Run(name, func(t *testing.T) {
			require.NotContains(t, basics, name)
			for _, elements := range steps {
				require.NotContains(t, elements, name)
			}
			require.NotContains(t, outputs, name)
			require.NotContains(t, rootParams, name)
			require.NotContains(t, wrapperParams, name)
		})
	}
	require.Equal(t, "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#", wrapperDoc["$schema"])

	loc := stackTemplateLocations(baseURL, stackVersion.AWSBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		DeploymentScope: string(app.StackDeploymentScopeSubscription),
		StackName:       "aca-medium",
		Region:          "eastus",
	})
	uiDefURL := baseURL + "/" + stackVersion.QuickLinkUIDefBucketKey
	require.Contains(t, loc.quickLinkURL, "/createUIDefinitionUri/")
	require.Contains(t, loc.quickLinkURL, escapeDataString(uiDefURL))
	require.NotContains(t, loc.quickLinkURL, ui.ContentsHash, "the portal link points at the rendered UI, not the vendor blob")

	rg := stackTemplateLocations(baseURL, stackVersion.AWSBucketKey, &CreateInstallStackVersionRequest{
		Platform:        string(app.AppRunnerTypeAzure),
		DeploymentScope: string(app.StackDeploymentScopeResourceGroup),
	})
	require.Empty(t, rg.quickLinkURL)
}

func acaMediumTemplateInput(cfg *config.AppConfig, ui *config.UploadedUIDefinition) *stacks.TemplateInput {
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "inst-aca-medium",
			AppID: "app-aca-medium",
			AzureAccount: &app.AzureAccount{
				Location:       "eastus",
				SubscriptionID: "00000000-1111-2222-3333-444444444444",
			},
		},
		CloudFormationStackVersion: &app.InstallStackVersion{
			PhoneHomeURL:            "https://api.example.com/phone-home/phid",
			AWSBucketKey:            "templates/ist-aca.json",
			QuickLinkBucketKey:      "templates/ist-aca-quicklink.json",
			QuickLinkUIDefBucketKey: "templates/ist-aca-ui.json",
		},
		Runner:   &app.Runner{ID: "runner-aca", OrgID: "org-aca"},
		Settings: &app.RunnerGroupSettings{ContainerImageURL: "example.com/runner", ContainerImageTag: "latest"},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				Type:                    app.StackTypeAzure,
				DeploymentScope:         app.StackDeploymentScopeSubscription,
				VPCNestedTemplateURL:    cfg.Stack.VPCNestedTemplateURL,
				RunnerNestedTemplateURL: cfg.Stack.RunnerNestedTemplateURL,
				UIDefinition:            ui,
			},
		},
		VPCNestedStackTemplateURL:    cfg.Stack.VPCNestedTemplateURL,
		RunnerNestedStackTemplateURL: cfg.Stack.RunnerNestedTemplateURL,
		DeploymentScope:              app.StackDeploymentScopeSubscription,
	}
	return inp
}

func indexUI(nodes []any) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, node := range nodes {
		el, ok := node.(map[string]any)
		if !ok {
			continue
		}
		name, _ := el["name"].(string)
		if name != "" {
			out[name] = el
		}
	}
	return out
}

func indexSteps(nodes []any) map[string]map[string]map[string]any {
	out := map[string]map[string]map[string]any{}
	for _, node := range nodes {
		step, ok := node.(map[string]any)
		if !ok {
			continue
		}
		name, _ := step["name"].(string)
		elements, _ := step["elements"].([]any)
		out[name] = indexUI(elements)
	}
	return out
}

type memoryBlobUploader struct {
	blobs map[string][]byte
}

func (m *memoryBlobUploader) UploadBlob(_ context.Context, blob []byte, key string) error {
	if m.blobs == nil {
		m.blobs = map[string][]byte{}
	}
	copied := make([]byte, len(blob))
	copy(copied, blob)
	m.blobs[key] = copied
	return nil
}
