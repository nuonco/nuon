package arm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const testTemplateURL = "https://templates.example.com/templates/test-install-id-00000001/ist-test.json"

func renderWrapper(t *testing.T, inp *stacks.TemplateInput) map[string]any {
	t.Helper()

	tmpl := &Templates{cfg: &internal.Config{}}
	byts, checksum, err := tmpl.QuickLinkWrapper(inp, testTemplateURL)
	if err != nil {
		t.Fatalf("QuickLinkWrapper returned error: %v", err)
	}
	if checksum == "" {
		t.Fatal("QuickLinkWrapper returned an empty checksum")
	}

	var out map[string]any
	if err := json.Unmarshal(byts, &out); err != nil {
		t.Fatalf("unable to unmarshal wrapper: %v", err)
	}
	return out
}

func wrapperStackResource(t *testing.T, wrapper map[string]any) (map[string]any, map[string]any) {
	t.Helper()

	resources, ok := wrapper["resources"].([]any)
	if !ok || len(resources) != 1 {
		t.Fatalf("expected exactly one wrapper resource, got %v", wrapper["resources"])
	}
	stack, ok := resources[0].(map[string]any)
	if !ok {
		t.Fatalf("wrapper resource is not an object: %v", resources[0])
	}
	props, ok := stack["properties"].(map[string]any)
	if !ok {
		t.Fatalf("stack resource missing properties: %v", stack)
	}
	return stack, props
}

func TestQuickLinkWrapper_StackResourceShape(t *testing.T) {
	wrapper := renderWrapper(t, minimalTemplateInput())
	stack, props := wrapperStackResource(t, wrapper)

	if got := stack["type"]; got != "Microsoft.Resources/deploymentStacks" {
		t.Errorf("unexpected resource type: %v", got)
	}
	if got := stack["apiVersion"]; got != deploymentStacksAPIVersion {
		t.Errorf("unexpected apiVersion: %v", got)
	}
	if got := stack["name"]; got != "test-install-id-00000001-stack" {
		t.Errorf("unexpected stack name: %v", got)
	}

	link, ok := props["templateLink"].(map[string]any)
	if !ok {
		t.Fatalf("stack missing templateLink: %v", props)
	}
	if got := link["uri"]; got != testTemplateURL {
		t.Errorf("unexpected templateLink uri: %v", got)
	}

	// The two protections a plain ARM deployment would lose.
	deny, ok := props["denySettings"].(map[string]any)
	if !ok {
		t.Fatalf("stack missing denySettings: %v", props)
	}
	if got := deny["mode"]; got != "denyDelete" {
		t.Errorf("unexpected denySettings mode: %v", got)
	}
	aou, ok := props["actionOnUnmanage"].(map[string]any)
	if !ok {
		t.Fatalf("stack missing actionOnUnmanage: %v", props)
	}
	if got := aou["resources"]; got != "delete" {
		t.Errorf("unexpected actionOnUnmanage.resources: %v", got)
	}
	if got := aou["resourcesWithoutDeleteSupport"]; got != "fail" {
		t.Errorf("unexpected actionOnUnmanage.resourcesWithoutDeleteSupport: %v", got)
	}
}

// A location on a resource-group-scoped deploymentStacks resource fails the
// deploy with InvalidTemplateDeployment, and `az deployment group validate`
// does not catch it — so the guard has to live here.
func TestQuickLinkWrapper_NoLocationAtResourceGroupScope(t *testing.T) {
	wrapper := renderWrapper(t, minimalTemplateInput())
	stack, _ := wrapperStackResource(t, wrapper)

	if _, present := stack["location"]; present {
		t.Errorf("resource-group scoped stack must not carry a location, got %v", stack["location"])
	}
	if got := wrapper["$schema"]; got != rgTemplateSchema {
		t.Errorf("unexpected schema: %v", got)
	}
}

func TestQuickLinkWrapper_LocationAtSubscriptionScope(t *testing.T) {
	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription

	wrapper := renderWrapper(t, inp)
	stack, _ := wrapperStackResource(t, wrapper)

	if got := stack["location"]; got != "eastus" {
		t.Errorf("subscription scoped stack must carry the install location, got %v", got)
	}
	if got := wrapper["$schema"]; got != subscriptionTemplateSchema {
		t.Errorf("unexpected schema: %v", got)
	}
}

// The portal builds its deployment form from the wrapper, so every parameter
// the stack template declares has to be re-declared and passed through —
// otherwise parameters without defaults (customer secrets) have no source.
func TestQuickLinkWrapper_ParametersArePassedThrough(t *testing.T) {
	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription
	inp.AppCfg.SecretsConfig.Secrets = []app.AppSecretConfig{
		{Name: "db_password", Required: true},
	}

	tmpl := &Templates{cfg: &internal.Config{}}
	inner, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	wrapper := renderWrapper(t, inp)
	_, props := wrapperStackResource(t, wrapper)

	wrapperParams, ok := wrapper["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("wrapper missing parameters: %v", wrapper)
	}
	passthrough, ok := props["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("stack missing parameters: %v", props)
	}

	if len(wrapperParams) != len(inner.Parameters) {
		t.Errorf("wrapper declares %d parameters, stack template declares %d", len(wrapperParams), len(inner.Parameters))
	}
	for name := range inner.Parameters {
		if _, present := wrapperParams[name]; !present {
			t.Errorf("wrapper does not re-declare parameter %q", name)
		}
		entry, present := passthrough[name].(map[string]any)
		if !present {
			t.Errorf("stack does not pass through parameter %q", name)
			continue
		}
		want := "[parameters('" + name + "')]"
		if entry["value"] != want {
			t.Errorf("parameter %q passed as %v, want %v", name, entry["value"], want)
		}
	}
}

// deployTimestamp defaults to [utcNow()], which ARM only accepts in a top-level
// template's parameter defaults. It stays a wrapper parameter for that reason.
func TestQuickLinkWrapper_DeployTimestampStaysAParameter(t *testing.T) {
	wrapper := renderWrapper(t, minimalTemplateInput())

	params, ok := wrapper["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("wrapper missing parameters: %v", wrapper)
	}
	ts, ok := params["deployTimestamp"].(map[string]any)
	if !ok {
		t.Fatalf("wrapper missing deployTimestamp parameter: %v", params)
	}
	if got := ts["defaultValue"]; got != "[utcNow()]" {
		t.Errorf("unexpected deployTimestamp default: %v", got)
	}
}

// Reprovision mints a new stack version, and therefore a new wrapper at a new
// URL. Deploying it must UPDATE the install's stack rather than create a second
// one, which holds only while the stack name derives from the install and not
// from the version. Anything that makes the name version-specific silently turns
// every reprovision into a parallel stack.
func TestQuickLinkWrapper_StackNameIsStableAcrossVersions(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	first := minimalTemplateInput()
	first.CloudFormationStackVersion.ID = "ist-version-one"
	second := minimalTemplateInput()
	second.CloudFormationStackVersion.ID = "ist-version-two"

	names := make([]string, 0, 2)
	for i, inp := range []*stacks.TemplateInput{first, second} {
		byts, _, err := tmpl.QuickLinkWrapper(inp, fmt.Sprintf("%s?v=%d", testTemplateURL, i))
		if err != nil {
			t.Fatalf("QuickLinkWrapper returned error: %v", err)
		}
		var out map[string]any
		if err := json.Unmarshal(byts, &out); err != nil {
			t.Fatalf("unable to unmarshal wrapper: %v", err)
		}
		stack, _ := wrapperStackResource(t, out)
		names = append(names, stack["name"].(string))
	}

	if names[0] != names[1] {
		t.Errorf("stack name changed between versions: %q vs %q", names[0], names[1])
	}
}

// The portal and the documented `az stack group create` command have to address
// the same stack, or a customer who switches between them ends up with two.
func TestQuickLinkWrapper_StackNameMatchesCLIConvention(t *testing.T) {
	inp := minimalTemplateInput()
	wrapper := renderWrapper(t, inp)
	stack, _ := wrapperStackResource(t, wrapper)

	if got, want := stack["name"], inp.Install.ID+"-stack"; got != want {
		t.Errorf("stack name = %v, want %v (the name used by az stack group create)", got, want)
	}
}

func TestQuickLinkWrapper_RequiresTemplateURL(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	if _, _, err := tmpl.QuickLinkWrapper(minimalTemplateInput(), ""); err == nil {
		t.Fatal("expected an error when the template URL is empty")
	}
}

func TestQuickLinkWrapper_NoNestedBrackets(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	byts, _, err := tmpl.QuickLinkWrapper(minimalTemplateInput(), testTemplateURL)
	if err != nil {
		t.Fatalf("QuickLinkWrapper returned error: %v", err)
	}
	assertNoNestedBrackets(t, byts)
}

func TestQuickLinkWrapper_ChecksumIsDeterministic(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	_, first, err := tmpl.QuickLinkWrapper(minimalTemplateInput(), testTemplateURL)
	if err != nil {
		t.Fatalf("QuickLinkWrapper returned error: %v", err)
	}
	_, second, err := tmpl.QuickLinkWrapper(minimalTemplateInput(), testTemplateURL)
	if err != nil {
		t.Fatalf("QuickLinkWrapper returned error: %v", err)
	}
	if first != second {
		t.Errorf("checksum is not deterministic: %q vs %q", first, second)
	}
}
