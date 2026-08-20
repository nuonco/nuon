package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func renderUIDef(t *testing.T, inp *stacks.TemplateInput) (map[string]any, map[string]any) {
	t.Helper()

	tmpl := &Templates{cfg: &internal.Config{}}
	byts, checksum, err := tmpl.QuickLinkUIDefinition(inp)
	if err != nil {
		t.Fatalf("QuickLinkUIDefinition returned error: %v", err)
	}
	if checksum == "" {
		t.Fatal("QuickLinkUIDefinition returned an empty checksum")
	}

	var out map[string]any
	if err := json.Unmarshal(byts, &out); err != nil {
		t.Fatalf("unable to unmarshal UI definition: %v", err)
	}
	params, ok := out["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("UI definition missing parameters: %v", out)
	}
	return out, params
}

func TestQuickLinkUIDefinition_Envelope(t *testing.T) {
	out, params := renderUIDef(t, minimalTemplateInput())

	if got := out["handler"]; got != "Microsoft.Azure.CreateUIDef" {
		t.Errorf("unexpected handler: %v", got)
	}
	// The version has to match the version embedded in $schema.
	if got := out["version"]; got != "0.1.2-preview" {
		t.Errorf("unexpected version: %v", got)
	}
	if schema, _ := out["$schema"].(string); !strings.Contains(schema, "0.1.2-preview") {
		t.Errorf("schema version does not match version field: %v", out["$schema"])
	}
	for _, key := range []string{"config", "basics", "steps", "outputs"} {
		if _, present := params[key]; !present {
			t.Errorf("UI definition missing parameters.%s", key)
		}
	}
}

func resourceGroupConfig(t *testing.T, params map[string]any) map[string]any {
	t.Helper()

	basics := params["config"].(map[string]any)["basics"].(map[string]any)
	rg, ok := basics["resourceGroup"].(map[string]any)
	if !ok {
		t.Fatalf("config.basics missing resourceGroup: %v", basics)
	}
	return rg
}

// On the first deploy the customer names the group. Constraining it here would
// reject a perfectly valid choice before the install has any group at all.
func TestQuickLinkUIDefinition_ResourceGroupUnconstrainedBeforeFirstDeploy(t *testing.T) {
	_, params := renderUIDef(t, minimalTemplateInput())
	rg := resourceGroupConfig(t, params)

	if _, present := rg["constraints"]; present {
		t.Errorf("resource group is constrained before the stack has phoned home: %v", rg["constraints"])
	}
	// The portal otherwise demands a new or empty group, which stops a customer
	// deploying into a group they already keep resources in.
	if rg["allowExisting"] != true {
		t.Errorf("resourceGroup.allowExisting = %v, want true", rg["allowExisting"])
	}
}

// Once the stack has reported where it landed, every later deploy has to go to
// the same group or it creates a second stack instead of updating this install.
func TestQuickLinkUIDefinition_PinsResourceGroupAfterFirstDeploy(t *testing.T) {
	inp := minimalTemplateInput()
	inp.InstallState = &state.State{
		InstallStack: &state.InstallStackState{
			Outputs: map[string]any{"resource_group_name": "customer-chosen-rg"},
		},
	}

	_, params := renderUIDef(t, inp)
	rg := resourceGroupConfig(t, params)

	validations := rg["constraints"].(map[string]any)["validations"].([]any)
	if len(validations) == 0 {
		t.Fatal("resourceGroup has no validations")
	}
	// Pinned to the group the customer actually used, NOT to <install-id>-rg.
	wantExpr := "[equals(resourceGroup().name, 'customer-chosen-rg')]"
	if got := validations[0].(map[string]any)["isValid"]; got != wantExpr {
		t.Errorf("isValid = %v, want %v", got, wantExpr)
	}
	if rg["allowExisting"] != true {
		t.Errorf("resourceGroup.allowExisting = %v, want true", rg["allowExisting"])
	}
}

// At subscription scope the stack template creates the resource group itself, so
// there is no picker to constrain and a validation would reject every deploy.
func TestQuickLinkUIDefinition_NoResourceGroupConstraintAtSubscriptionScope(t *testing.T) {
	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription

	_, params := renderUIDef(t, inp)
	basics := params["config"].(map[string]any)["basics"].(map[string]any)

	if _, present := basics["resourceGroup"]; present {
		t.Errorf("subscription scope must not constrain the resource group: %v", basics["resourceGroup"])
	}
}

func TestQuickLinkUIDefinition_PinsLocation(t *testing.T) {
	_, params := renderUIDef(t, minimalTemplateInput())
	basics := params["config"].(map[string]any)["basics"].(map[string]any)

	location, ok := basics["location"].(map[string]any)
	if !ok {
		t.Fatalf("config.basics missing location: %v", basics)
	}
	allowed, ok := location["allowedValues"].([]any)
	if !ok || len(allowed) != 1 || allowed[0] != "eastus" {
		t.Errorf("location.allowedValues = %v, want [eastus]", location["allowedValues"])
	}
}

func subscriptionValidations(t *testing.T, params map[string]any) []any {
	t.Helper()

	basics := params["config"].(map[string]any)["basics"].(map[string]any)
	return basics["subscription"].(map[string]any)["constraints"].(map[string]any)["validations"].([]any)
}

func TestQuickLinkUIDefinition_RequiresStackWritePermission(t *testing.T) {
	_, params := renderUIDef(t, minimalTemplateInput())

	found := false
	for _, v := range subscriptionValidations(t, params) {
		if v.(map[string]any)["permission"] == "Microsoft.Resources/deploymentStacks/write" {
			found = true
		}
	}
	if !found {
		t.Errorf("subscription validations do not require deploymentStacks/write")
	}
}

// Scope is subscription + resource group. Pinning only the group still lets a
// deploy into another subscription create a second stack.
func TestQuickLinkUIDefinition_PinsSubscriptionWhenKnown(t *testing.T) {
	inp := minimalTemplateInput()
	inp.Install.AzureAccount.SubscriptionID = "00000000-1111-2222-3333-444444444444"

	_, params := renderUIDef(t, inp)

	want := "[equals(subscription().subscriptionId, '00000000-1111-2222-3333-444444444444')]"
	found := false
	for _, v := range subscriptionValidations(t, params) {
		if v.(map[string]any)["isValid"] == want {
			found = true
		}
	}
	if !found {
		t.Errorf("subscription is not pinned to the install's subscription: %v", subscriptionValidations(t, params))
	}
}

// The subscription is only mandatory for orgs with phone-home auth on. Emitting
// an equality check against an empty string would reject every subscription.
func TestQuickLinkUIDefinition_NoSubscriptionPinWhenUnknown(t *testing.T) {
	inp := minimalTemplateInput()
	inp.Install.AzureAccount.SubscriptionID = ""

	_, params := renderUIDef(t, inp)

	for _, v := range subscriptionValidations(t, params) {
		if _, present := v.(map[string]any)["isValid"]; present {
			t.Errorf("emitted a subscription equality check with no subscription to pin: %v", v)
		}
	}
}

// Every parameter the wrapper declares needs a field, whether or not it carries a
// default — a defaulted parameter omitted here is one the customer can never
// change, and an app whose parameters all have defaults gets a Basics step with
// nothing on it but subscription and region. Secrets must not render as plain text
// boxes.
func TestQuickLinkUIDefinition_PromptsForEveryParameter(t *testing.T) {
	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription
	inp.AppCfg.SecretsConfig.Secrets = []app.AppSecretConfig{
		{Name: "db_password", Required: true},
	}

	tmpl := &Templates{cfg: &internal.Config{}}
	wrapperParams, err := tmpl.quickLinkWrapperParameters(inp)
	if err != nil {
		t.Fatalf("quickLinkWrapperParameters returned error: %v", err)
	}

	_, params := renderUIDef(t, inp)
	basics := params["basics"].([]any)
	outputs := params["outputs"].(map[string]any)

	byName := map[string]map[string]any{}
	for _, b := range basics {
		el := b.(map[string]any)
		byName[el["name"].(string)] = el
	}

	for name, p := range wrapperParams {
		// location comes from the Basics step's own region picker, and
		// deployTimestamp defaults to [utcNow()], which only ARM can evaluate —
		// as a field it would show the customer the literal expression.
		if name == "location" || name == "deployTimestamp" {
			if _, present := byName[name]; present {
				t.Errorf("Nuon-managed parameter %q is exposed as a field", name)
			}
			continue
		}
		el, present := byName[name]
		if !present {
			t.Errorf("parameter %q is not prompted for", name)
			continue
		}
		if p.Type == "securestring" && el["type"] != "Microsoft.Common.PasswordBox" {
			t.Errorf("securestring parameter %q rendered as %v, want a PasswordBox", name, el["type"])
		}
		if got, want := outputs[name], "[basics('"+name+"')]"; got != want {
			t.Errorf("outputs[%q] = %v, want %v", name, got, want)
		}
	}
}

// The regression that prompted this: a custom VNet template hoists its CIDR
// parameters into the root, all with defaults. Filtering defaulted parameters out
// left the portal form empty while deploying the stack template directly showed
// every one of them.
func TestQuickLinkUIDefinition_PrefillsHoistedParametersWithTheirDefaults(t *testing.T) {
	inp := vnetInputWithTemplate(t, app.StackDeploymentScopeSubscription, hoistFixture)

	_, params := renderUIDef(t, inp)

	byName := map[string]map[string]any{}
	for _, b := range params["basics"].([]any) {
		el := b.(map[string]any)
		byName[el["name"].(string)] = el
	}

	addressSpace, present := byName["addressSpace"]
	if !present {
		t.Fatalf("hoisted parameter addressSpace is not prompted for; basics = %v", byName)
	}
	if got := addressSpace["defaultValue"]; got != "10.100.0.0/22" {
		t.Errorf("addressSpace.defaultValue = %v, want the template's default", got)
	}
	// The portal spaces and title-cases parameter names itself when no UI
	// definition is supplied; supplying one takes that over.
	if got := addressSpace["label"]; got != "Address Space" {
		t.Errorf("addressSpace.label = %v, want %q", got, "Address Space")
	}
	// Clearing a prefilled field must not submit an empty string over the default.
	if got := addressSpace["constraints"].(map[string]any)["required"]; got != true {
		t.Errorf("addressSpace.constraints.required = %v, want true", got)
	}

	if got := byName["peeringEnabled"]["type"]; got != "Microsoft.Common.CheckBox" {
		t.Errorf("bool parameter rendered as %v, want a CheckBox", got)
	}
}

func TestHumanizeParamName(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"addressSpace", "Address Space"},
		{"vnetCIDR", "Vnet CIDR"},
		{"secretTailscaleOauthClientId", "Secret Tailscale Oauth Client Id"},
		{"location", "Location"},
		{"", ""},
	} {
		if got := humanizeParamName(tc.in); got != tc.want {
			t.Errorf("humanizeParamName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Outputs may only name parameters the wrapper actually declares; the portal
// rejects a deployment whose outputs reference an unknown parameter.
func TestQuickLinkUIDefinition_OutputsMatchWrapperParameters(t *testing.T) {
	inp := minimalTemplateInput()

	tmpl := &Templates{cfg: &internal.Config{}}
	wrapperParams, err := tmpl.quickLinkWrapperParameters(inp)
	if err != nil {
		t.Fatalf("quickLinkWrapperParameters returned error: %v", err)
	}

	_, params := renderUIDef(t, inp)
	for name := range params["outputs"].(map[string]any) {
		if _, declared := wrapperParams[name]; !declared {
			t.Errorf("outputs references %q, which the wrapper does not declare", name)
		}
	}
}

func TestQuickLinkUIDefinition_ChecksumIsDeterministic(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	_, first, err := tmpl.QuickLinkUIDefinition(minimalTemplateInput())
	if err != nil {
		t.Fatalf("QuickLinkUIDefinition returned error: %v", err)
	}
	_, second, err := tmpl.QuickLinkUIDefinition(minimalTemplateInput())
	if err != nil {
		t.Fatalf("QuickLinkUIDefinition returned error: %v", err)
	}
	if first != second {
		t.Errorf("checksum is not deterministic: %q vs %q", first, second)
	}
}
