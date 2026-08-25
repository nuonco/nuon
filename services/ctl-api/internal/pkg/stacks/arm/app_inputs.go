package arm

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// installInputsEnvName carries the whole install_inputs object into the phone-home
// script as a single JSON blob rather than one env var per input — see
// installInputsObjectExpr for why that distinction is load-bearing.
const installInputsEnvName = "INSTALL_INPUTS_JSON"

// camelParamName builds an ARM parameter name from a config-supplied name. The portal
// derives its form label from the parameter name, so this camelCases rather than
// carrying the config's underscores through: db_password reads as "Db Password".
func camelParamName(prefix, name string) string {
	out := prefix
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		r := []rune(part)
		out += strings.ToUpper(string(r[0])) + string(r[1:])
	}
	return out
}

// azureInputParamName turns a customer-source app input name into an ARM parameter
// name.
//
// The prefix is not cosmetic: an input name is vendor-chosen, so without it an input
// called `location` would collide with a Nuon-managed parameter and an unlucky one
// could shadow a parameter hoisted off a custom VNet template.
func azureInputParamName(name string) string {
	return camelParamName("input", name)
}

type azureInput struct {
	name        string
	paramName   string
	label       string
	description string

	// value is what the portal pre-fills and the CLI defaults to.
	//
	// Seeded from the install's current inputs rather than the app's declared
	// default because the phone-home reports every parameter back as the install's
	// inputs: defaulting to the vendor's default would revert a value the customer
	// set through the dashboard on the next reprovision.
	value string

	required bool
}

// azureCustomerInputs are the app inputs the customer supplies at deploy time.
// Sorted so the render is deterministic.
//
// Reads InputConfig.AppInputs alone, which carries every declared input, grouped ones
// included — AppInputGroups.AppInputs is not loaded on the config the renderer
// receives, and the CloudFormation renderer partitions the same flat list by group.
func azureCustomerInputs(inp *stacks.TemplateInput) []azureInput {
	if inp == nil || inp.AppCfg == nil {
		return nil
	}

	current := map[string]*string{}
	if inp.Install != nil && inp.Install.CurrentInstallInputs != nil {
		current = inp.Install.CurrentInstallInputs.Values
	}

	var out []azureInput
	for _, in := range inp.AppCfg.InputConfig.AppInputs {
		if in.Source != app.AppInputSourceCustomer {
			continue
		}

		// Blank counts as unset, matching how the install's input state resolves a
		// materialized-but-empty value back to the declared default.
		value := in.Default
		if v := generics.FromPtrStr(current[in.Name]); v != "" {
			value = v
		}

		label := in.DisplayName
		if label == "" {
			label = humanizeParamName(camelParamName("", in.Name))
		}

		out = append(out, azureInput{
			name:        in.Name,
			paramName:   azureInputParamName(in.Name),
			label:       label,
			description: in.Description,
			value:       value,
			required:    in.Required,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })

	return out
}

// azureInputParameter surfaces one customer input as a root parameter, which the
// portal renders as a field on the deployment form and `az stack ... --parameters`
// sets from the command line.
//
// Always a plain string, whatever the input's declared type. Install inputs are
// stored and validated as strings by the control plane, and ARM's own scalar types
// would both rewrite and reject legitimate values on the way through: string(true)
// is "True", not "true", and an int parameter refuses the fractional value an app
// input of type number is allowed to hold.
//
// Sensitive inputs are not securestring. A securestring's value still has to reach
// the phone-home script's environment to be reported back, where it is readable by
// anyone with read access on the resource group — a masked field over a value that
// is not actually protected is worse than an honest one. Secrets belong in the app's
// secrets config, which renders as securestring into Key Vault.
func azureInputParameter(in azureInput) ARMParameter {
	p := ARMParameter{Type: "string"}

	// A required input with nothing to prefill gets no default at all: that is what
	// makes ARM refuse the deploy rather than write a blank value.
	if !in.required || in.value != "" {
		p.DefaultValue = in.value
	}
	if in.description != "" {
		p.Metadata = &ARMParameterMetadata{Description: in.description}
	}

	return p
}

// addCustomerInputParameters declares the customer inputs on the root template.
//
// Called after every other parameter source has written, so a name that collides
// with one of theirs is an error rather than a silent overwrite — the overwrite would
// render a template that deploys and then reports the wrong value home.
func addCustomerInputParameters(tmpl *ARMTemplate, inp *stacks.TemplateInput) error {
	claimed := map[string]string{}
	for _, in := range azureCustomerInputs(inp) {
		if other, dup := claimed[in.paramName]; dup {
			return fmt.Errorf("customer inputs %q and %q both map to ARM parameter %q", other, in.name, in.paramName)
		}
		if _, taken := tmpl.Parameters[in.paramName]; taken {
			return fmt.Errorf("customer input %q maps to ARM parameter %q, which is already declared", in.name, in.paramName)
		}
		if slices.Contains(ReservedParamNames, in.paramName) {
			return fmt.Errorf("customer input %q maps to reserved ARM parameter %q", in.name, in.paramName)
		}

		claimed[in.paramName] = in.name
		tmpl.Parameters[in.paramName] = azureInputParameter(in)
	}

	return nil
}

// azureInputLabels is the portal field label for each customer-input parameter, keyed
// by parameter name, so the customer reads the vendor's display name rather than the
// name the renderer derived from the input.
func azureInputLabels(inp *stacks.TemplateInput) map[string]string {
	labels := map[string]string{}
	for _, in := range azureCustomerInputs(inp) {
		labels[in.paramName] = in.label
	}
	return labels
}

// installInputsObjectExpr is the install_inputs object as an ARM expression that
// evaluates to its JSON text.
//
// The escaping is the point. Every other field in the phone-home payload is
// interpolated into the heredoc as `"key": "$VAR"`, which holds only because those
// values are Azure resource IDs. An input value is whatever the customer typed, so a
// quote, backslash or newline in one would produce a malformed body and fail the
// deploy on a valid input. ARM's string() escapes the values for us, and the result
// is spliced into the payload unquoted because it is already a JSON object.
func installInputsObjectExpr(inputs []azureInput) string {
	args := make([]string, 0, len(inputs)*2)
	for _, in := range inputs {
		args = append(args, armStringLiteral(in.name), fmt.Sprintf("parameters('%s')", in.paramName))
	}
	return fmt.Sprintf("[string(createObject(%s))]", strings.Join(args, ", "))
}

// armStringLiteral quotes a value as an ARM string literal, where a single quote is
// escaped by doubling it. Needed here and not elsewhere in the package because these
// keys are vendor-chosen input names rather than renderer-owned identifiers.
func armStringLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
