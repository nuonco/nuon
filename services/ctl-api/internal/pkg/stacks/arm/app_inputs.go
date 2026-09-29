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

const installInputsEnvName = "INSTALL_INPUTS_JSON"

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

func azureInputParamName(name string) string {
	return camelParamName("input", name)
}

type azureInput struct {
	name        string
	paramName   string
	label       string
	description string

	value string

	required bool
}

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

func azureInputParameter(in azureInput) ARMParameter {
	p := ARMParameter{Type: "string"}

	if !in.required || in.value != "" {
		p.DefaultValue = in.value
	}
	if in.description != "" {
		p.Metadata = &ARMParameterMetadata{Description: in.description}
	}

	return p
}

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

func azureInputLabels(inp *stacks.TemplateInput) map[string]string {
	labels := map[string]string{}
	for _, in := range azureCustomerInputs(inp) {
		labels[in.paramName] = in.label
	}
	return labels
}

// why: installInputsObjectExpr is the install_inputs object as an ARM expression that
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

// why: armStringLiteral quotes a value as an ARM string literal, where a single quote is
// escaped by doubling it. Needed here and not elsewhere in the package because these
// keys are vendor-chosen input names rather than renderer-owned identifiers.
func armStringLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
