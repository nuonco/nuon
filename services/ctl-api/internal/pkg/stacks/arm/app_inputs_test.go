package arm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func customerInputsTemplateInput(inputs ...app.AppInput) *stacks.TemplateInput {
	inp := minimalTemplateInput()
	inp.AppCfg.InputConfig.AppInputs = inputs
	return inp
}

func customerInput(name string) app.AppInput {
	return app.AppInput{Name: name, Source: app.AppInputSourceCustomer}
}

func TestCustomerInputs_RootParameters(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	dbName := customerInput("db_name")
	dbName.Default = "postgres"
	dbName.Description = "The database to connect to."

	replicas := customerInput("replica-count")
	replicas.Type = app.AppInputTypeNumber
	replicas.Default = "3"

	apiKey := customerInput("api_key")
	apiKey.Required = true

	optional := customerInput("log_level")

	vendorOwned := app.AppInput{Name: "internal_thing", Source: app.AppInputSourceVendor, Default: "x"}

	armTmpl, err := tmpl.getAzureTemplate(customerInputsTemplateInput(
		dbName, replicas, apiKey, optional, vendorOwned,
	))
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	if _, present := armTmpl.Parameters["inputInternalThing"]; present {
		t.Error("vendor-source input must not be exposed as a template parameter")
	}

	for _, tc := range []struct {
		param    string
		wantType string
		// wantDefault nil asserts the parameter carries no defaultValue at all,
		// which is what makes ARM refuse the deploy without a value.
		wantDefault any
		wantDesc    string
	}{
		{param: "inputDbName", wantType: "string", wantDefault: "postgres", wantDesc: "The database to connect to."},
		// A number input is still a string parameter: ARM's int would reject the
		// fractional values an app input of type number may hold.
		{param: "inputReplicaCount", wantType: "string", wantDefault: "3"},
		{param: "inputApiKey", wantType: "string", wantDefault: nil},
		// Optional with no declared default: an empty default keeps the deploy
		// possible without a value.
		{param: "inputLogLevel", wantType: "string", wantDefault: ""},
	} {
		p, present := armTmpl.Parameters[tc.param]
		if !present {
			t.Errorf("template missing parameter %s", tc.param)
			continue
		}
		if p.Type != tc.wantType {
			t.Errorf("%s type = %q, want %q", tc.param, p.Type, tc.wantType)
		}
		if p.DefaultValue != tc.wantDefault {
			t.Errorf("%s defaultValue = %#v, want %#v", tc.param, p.DefaultValue, tc.wantDefault)
		}
		desc := ""
		if p.Metadata != nil {
			desc = p.Metadata.Description
		}
		if desc != tc.wantDesc {
			t.Errorf("%s description = %q, want %q", tc.param, desc, tc.wantDesc)
		}
	}

	tmplBytes, err := json.MarshalIndent(armTmpl, "", "  ")
	if err != nil {
		t.Fatalf("unable to marshal ARM template: %v", err)
	}
	assertNoNestedBrackets(t, tmplBytes)

	// omitempty on an interface field only drops a nil, so an empty default has to
	// survive the marshal — dropping it would make an optional input mandatory.
	if !strings.Contains(string(tmplBytes), `"inputLogLevel": {`) ||
		!strings.Contains(string(tmplBytes), `"defaultValue": ""`) {
		t.Errorf("optional input lost its empty defaultValue:\n%s", tmplBytes)
	}
}

// The phone home reports every parameter back as the install's inputs, so a
// parameter defaulting to the vendor's default would revert a value the customer
// set through the dashboard on the next reprovision.
func TestCustomerInputs_CurrentInstallValueWins(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	declared := customerInput("db_name")
	declared.Default = "postgres"
	blanked := customerInput("log_level")
	blanked.Default = "info"

	inp := customerInputsTemplateInput(declared, blanked)
	inp.Install.CurrentInstallInputs = &app.InstallInputs{
		Values: pgtype.Hstore{
			"db_name": generics.ToPtr("orders"),
			// Materialized but blank counts as unset, matching how the install's
			// input state resolves one back to the declared default.
			"log_level": generics.ToPtr(""),
		},
	}

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	if got := armTmpl.Parameters["inputDbName"].DefaultValue; got != "orders" {
		t.Errorf("inputDbName defaultValue = %#v, want %q", got, "orders")
	}
	if got := armTmpl.Parameters["inputLogLevel"].DefaultValue; got != "info" {
		t.Errorf("inputLogLevel defaultValue = %#v, want %q", got, "info")
	}
}

func TestCustomerInputs_PhoneHomePayload(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	res := phoneHomeScript(t, tmpl, customerInputsTemplateInput(
		customerInput("db_name"), customerInput("api_key"),
	), nil)

	props := res["properties"].(map[string]any)
	script := props["scriptContent"].(string)
	// Unquoted: the env var already holds a JSON object, so quoting it would send a
	// string where the endpoint decodes a map.
	if !strings.Contains(script, `"install_inputs": $INSTALL_INPUTS_JSON`) {
		t.Errorf("payload missing install_inputs:\n%s", script)
	}

	var got string
	for _, env := range props["environmentVariables"].([]map[string]any) {
		if env["name"] == installInputsEnvName {
			got = env["value"].(string)
		}
	}
	// Sorted by input name, so the render is stable across reprovisions.
	want := "[string(createObject('api_key', parameters('inputApiKey'), 'db_name', parameters('inputDbName')))]"
	if got != want {
		t.Errorf("install_inputs env value = %q, want %q", got, want)
	}

	resBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	assertNoNestedBrackets(t, resBytes)
}

// An app that declares no customer inputs must not gain the field: the goldens
// assert the rest of the payload, and an empty env var would splice invalid JSON.
func TestCustomerInputs_PhoneHomePayloadOmittedWhenNone(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	res := phoneHomeScript(t, tmpl, minimalTemplateInput(), nil)

	script := res["properties"].(map[string]any)["scriptContent"].(string)
	if strings.Contains(script, "install_inputs") {
		t.Errorf("payload should not mention install_inputs:\n%s", script)
	}
}

// ARM escapes a single quote by doubling it. Unescaped, an input name carrying one
// would truncate the createObject argument list into a template that fails preflight.
func TestCustomerInputs_PhoneHomeQuotesInputNames(t *testing.T) {
	got := installInputsObjectExpr([]azureInput{{name: "it's_fine", paramName: "inputItSFine"}})
	want := "[string(createObject('it''s_fine', parameters('inputItSFine')))]"
	if got != want {
		t.Errorf("expression = %q, want %q", got, want)
	}
}

func TestCustomerInputs_ParameterCollisions(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	t.Run("two inputs mapping to one parameter", func(t *testing.T) {
		_, err := tmpl.getAzureTemplate(customerInputsTemplateInput(
			customerInput("db_name"), customerInput("db-name"),
		))
		if err == nil {
			t.Fatal("expected an error for inputs colliding on one ARM parameter")
		}
		if !strings.Contains(err.Error(), "inputDbName") {
			t.Errorf("error does not name the parameter: %v", err)
		}
	})

	// The realistic shadowing case is a parameter hoisted off a custom VNet or nested
	// stack template that happens to be named like a derived input one. A silent
	// overwrite there would deploy and then report the wrong value home, so the
	// adder is exercised against an already-populated parameter map directly.
	t.Run("input shadowing a parameter another source owns", func(t *testing.T) {
		existing := &ARMTemplate{
			Parameters: map[string]ARMParameter{
				"inputDbName": {Type: "string", DefaultValue: "hoisted"},
			},
		}

		err := addCustomerInputParameters(existing, customerInputsTemplateInput(customerInput("db_name")))
		if err == nil {
			t.Fatal("expected an error for an input colliding with an existing parameter")
		}
		if got := existing.Parameters["inputDbName"].DefaultValue; got != "hoisted" {
			t.Errorf("existing parameter was overwritten: %#v", got)
		}
	})
}

// The prefix is what keeps a vendor-chosen input name off the Nuon-managed
// parameters, so it has to hold for the names most likely to collide.
func TestCustomerInputs_ReservedNamesAreUnreachable(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	var inputs []app.AppInput
	for _, name := range ReservedParamNames {
		inputs = append(inputs, customerInput(name))
	}

	armTmpl, err := tmpl.getAzureTemplate(customerInputsTemplateInput(inputs...))
	if err != nil {
		t.Fatalf("getAzureTemplate returned error: %v", err)
	}

	if got := armTmpl.Parameters["location"].DefaultValue; got != "eastus" {
		t.Errorf("location parameter was clobbered: %#v", got)
	}
	if _, present := armTmpl.Parameters["inputLocation"]; !present {
		t.Error("an input named location should render as inputLocation")
	}
}

func TestCustomerInputs_QuickLinkBasics(t *testing.T) {
	labelled := customerInput("db_name")
	labelled.DisplayName = "Database"
	labelled.Default = "postgres"
	labelled.Description = "The database to connect to."

	unlabelled := customerInput("replica_count")
	unlabelled.Default = "3"

	required := customerInput("api_key")
	required.Required = true

	optional := customerInput("log_level")

	inp := customerInputsTemplateInput(labelled, unlabelled, required, optional)
	inp.DeploymentScope = app.StackDeploymentScopeSubscription

	_, params := renderUIDef(t, inp)

	byName := map[string]map[string]any{}
	for _, element := range params["basics"].([]any) {
		el := element.(map[string]any)
		byName[el["name"].(string)] = el
	}

	for _, tc := range []struct {
		param        string
		wantLabel    string
		wantRequired bool
		wantDefault  any
		wantToolTip  string
	}{
		// The vendor's display name, so the portal reads the same as the dashboard.
		{param: "inputDbName", wantLabel: "Database", wantRequired: true, wantDefault: "postgres", wantToolTip: "The database to connect to."},
		{param: "inputReplicaCount", wantLabel: "Replica Count", wantRequired: true, wantDefault: "3"},
		{param: "inputApiKey", wantLabel: "Api Key", wantRequired: true, wantDefault: nil},
		// Blank is a legitimate answer here, so requiring it would leave the form
		// unsubmittable.
		{param: "inputLogLevel", wantLabel: "Log Level", wantRequired: false, wantDefault: ""},
	} {
		el, present := byName[tc.param]
		if !present {
			t.Errorf("basics step missing %s", tc.param)
			continue
		}
		if el["label"] != tc.wantLabel {
			t.Errorf("%s label = %v, want %q", tc.param, el["label"], tc.wantLabel)
		}
		if el["defaultValue"] != tc.wantDefault {
			t.Errorf("%s defaultValue = %#v, want %#v", tc.param, el["defaultValue"], tc.wantDefault)
		}
		if el["toolTip"] != tc.wantToolTip {
			t.Errorf("%s toolTip = %v, want %q", tc.param, el["toolTip"], tc.wantToolTip)
		}
		required, _ := el["constraints"].(map[string]any)["required"].(bool)
		if required != tc.wantRequired {
			t.Errorf("%s required = %v, want %v", tc.param, required, tc.wantRequired)
		}

		outputs := params["outputs"].(map[string]any)
		if got := outputs[tc.param]; got != "[basics('"+tc.param+"')]" {
			t.Errorf("%s output = %v", tc.param, got)
		}
	}
}

// The portal builds its form from the wrapper, so a parameter the wrapper does not
// re-declare and pass through is a field the customer fills in and the stack never
// sees.
func TestCustomerInputs_QuickLinkWrapperPassthrough(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	inp := customerInputsTemplateInput(customerInput("db_name"))
	inp.DeploymentScope = app.StackDeploymentScopeSubscription

	byts, _, err := tmpl.QuickLinkWrapper(inp, "https://example.com/template.json")
	if err != nil {
		t.Fatalf("QuickLinkWrapper returned error: %v", err)
	}

	var wrapper struct {
		Parameters map[string]ARMParameter `json:"parameters"`
		Resources  []struct {
			Properties struct {
				Parameters map[string]map[string]any `json:"parameters"`
			} `json:"properties"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(byts, &wrapper); err != nil {
		t.Fatalf("unmarshal wrapper: %v", err)
	}

	if _, present := wrapper.Parameters["inputDbName"]; !present {
		t.Error("wrapper does not re-declare inputDbName")
	}
	if got := wrapper.Resources[0].Properties.Parameters["inputDbName"]["value"]; got != "[parameters('inputDbName')]" {
		t.Errorf("wrapper does not pass inputDbName through: %v", got)
	}
}
