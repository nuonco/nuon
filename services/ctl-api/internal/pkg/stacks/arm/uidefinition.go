package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) QuickLinkUIDefinition(inp *stacks.TemplateInput) ([]byte, string, error) {
	wrapperParams, err := t.quickLinkWrapperParameters(inp)
	if err != nil {
		return nil, "", err
	}

	scope := scopeFor(inp)
	location := inp.Install.AzureAccount.Location

	// why: A deployment stack is identified by scope AND name, so pinning the name is
	// only half of it: deploying into a different subscription produces a second,
	// independent stack rather than updating the install's. The subscription is
	// captured when the install is created, but it is only mandatory for orgs with
	// phone-home auth enabled — where it is unknown, fall back to the permission
	// check alone rather than emitting a validation that can never pass.
	subscriptionValidations := []any{}
	if subID := inp.Install.AzureAccount.SubscriptionID; subID != "" {
		subscriptionValidations = append(subscriptionValidations, map[string]any{
			"isValid": fmt.Sprintf("[equals(subscription().subscriptionId, '%s')]", subID),
			"message": fmt.Sprintf("This install targets subscription %s. Deploying into a different subscription creates a second stack instead of updating this install.", subID),
		})
	}
	subscriptionValidations = append(subscriptionValidations, map[string]any{
		"permission": "Microsoft.Resources/deploymentStacks/write",
		"message":    "You need permission to create deployment stacks in this subscription.",
	})

	basicsConfig := map[string]any{
		"description": fmt.Sprintf(
			"Deploys the Nuon install stack for `%s`. Re-running this for an existing install updates its deployment stack in place.",
			inp.Install.ID,
		),
		"subscription": map[string]any{
			"constraints":       map[string]any{"validations": subscriptionValidations},
			"resourceProviders": []string{"Microsoft.Compute"},
		},
		"location": locationPin(location),
	}

	if !scope.subscription {
		resourceGroup := map[string]any{"allowExisting": true}
		if rgName := deployedResourceGroupName(inp); rgName != "" {
			resourceGroup["constraints"] = map[string]any{
				"validations": []any{
					map[string]any{
						"isValid": fmt.Sprintf("[equals(resourceGroup().name, '%s')]", rgName),
						"message": fmt.Sprintf("This install is deployed to %s. Deploying into a different resource group creates a second stack instead of updating this install.", rgName),
					},
				},
			}
		}
		basicsConfig["resourceGroup"] = resourceGroup
	}

	basics := []any{}
	outputs := map[string]any{}
	if _, declared := wrapperParams["location"]; declared {
		outputs["location"] = "[location()]"
	}

	inputLabels := azureInputLabels(inp)
	for _, name := range sortedParamNames(wrapperParams) {
		if name == "location" || name == "deployTimestamp" {
			continue
		}

		if name == runnerVmSizeParamName {
			basics = append(basics, runnerVMSizeUIElement(wrapperParams[name]))
			outputs[name] = "[basics('" + runnerVmSizeParamName + "')]"
			continue
		}

		element, output, ok := basicsElement(name, wrapperParams[name], inputLabels[name])
		if !ok {
			continue
		}
		basics = append(basics, element)
		outputs[name] = output
	}

	uiDef := map[string]any{
		"$schema": "https://schema.management.azure.com/schemas/0.1.2-preview/CreateUIDefinition.MultiVm.json#",
		"handler": "Microsoft.Azure.CreateUIDef",
		"version": "0.1.2-preview",
		"parameters": map[string]any{
			"config":  map[string]any{"basics": basicsConfig},
			"basics":  basics,
			"steps":   []any{},
			"outputs": outputs,
		},
	}

	uiDefBytes, err := json.MarshalIndent(uiDef, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal quick link UI definition: %w", err)
	}

	hash := sha256.Sum256(uiDefBytes)
	return uiDefBytes, hex.EncodeToString(hash[:]), nil
}

func runnerVMSizeUIElement(p ARMParameter) map[string]any {
	allowed := make([]string, 0, len(p.AllowedValues))
	for _, value := range p.AllowedValues {
		allowed = append(allowed, fmt.Sprintf("%v", value))
	}

	recommended := append([]string{}, allowed...)
	if def, ok := p.DefaultValue.(string); ok && def != "" {
		rest := make([]string, 0, len(allowed))
		for _, size := range allowed {
			if size != def {
				rest = append(rest, size)
			}
		}
		if len(allowed) == 0 {
			recommended = []string{def}
		} else {
			recommended = append([]string{def}, rest...)
		}
	}

	element := map[string]any{
		"name":             runnerVmSizeParamName,
		"type":             "Microsoft.Compute.SizeSelector",
		"label":            "Runner VM Size",
		"osPlatform":       "Linux",
		"count":            1,
		"recommendedSizes": recommended,
	}
	if p.Metadata != nil && p.Metadata.Description != "" {
		element["toolTip"] = p.Metadata.Description
	}
	if len(allowed) > 0 {
		element["constraints"] = map[string]any{"allowedSizes": allowed}
	}
	return element
}

func locationPin(location string) map[string]any {
	if location == "" {
		return map[string]any{}
	}
	return map[string]any{
		"allowedValues": []string{location},
		"toolTip":       "The install's region. It is fixed for the lifetime of the install.",
	}
}

func deployedResourceGroupName(inp *stacks.TemplateInput) string {
	if inp.InstallState == nil || inp.InstallState.InstallStack == nil {
		return ""
	}
	name, _ := inp.InstallState.InstallStack.Outputs["resource_group_name"].(string)
	return name
}

func sortedParamNames(params map[string]ARMParameter) []string {
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// why: basicsElement renders one wrapper parameter as a field on the Basics step,
// returning the element, the outputs expression that feeds its value back to the
// wrapper, and whether the parameter is renderable at all.
//
// Every parameter gets a field, including those carrying a default — the default
// becomes the field's initial value. Omitting defaulted parameters instead, as an
// earlier version did, silently dropped the customer's only chance to change a
// VNet address space or subnet CIDR, and left apps whose parameters all had
// defaults with a Basics step showing nothing but subscription and region.
//
// A field carrying a non-empty default is still marked required, so a customer who
// clears one cannot submit an empty string in place of the default. A default that
// is itself empty is the one case where blank is a legitimate answer — an optional
// app input the vendor declared no default for — so requiring it there would leave
// the form unsubmittable.
//
// label overrides the name-derived one, for parameters whose config carries a
// display name of its own. Empty falls back to the derived label.
//
// Object and array parameters have no sensible Basics element and are skipped, so
// the wrapper's own default applies. Nothing currently reaches the root with those
// types: a nested template's non-scalar default is either Nuon-managed or left
// unhoisted.
func basicsElement(name string, p ARMParameter, label string) (map[string]any, string, bool) {
	if label == "" {
		label = humanizeParamName(name)
	}
	element := map[string]any{
		"name":    name,
		"label":   label,
		"toolTip": "",
	}
	if p.Metadata != nil && p.Metadata.Description != "" {
		element["toolTip"] = p.Metadata.Description
	}

	if p.Type != "securestring" && len(p.AllowedValues) > 0 && (p.Type == "string" || p.Type == "int" || p.Type == "bool") {
		allowedValues := make([]any, 0, len(p.AllowedValues))
		for _, value := range p.AllowedValues {
			allowedValues = append(allowedValues, map[string]any{
				"label": humanizeParamName(fmt.Sprintf("%v", value)),
				"value": value,
			})
		}
		element["type"] = "Microsoft.Common.DropDown"
		element["constraints"] = map[string]any{
			"allowedValues": allowedValues,
			"required":      true,
		}
		if p.DefaultValue != nil {
			element["defaultValue"] = humanizeParamName(fmt.Sprintf("%v", p.DefaultValue))
		}
		return element, fmt.Sprintf("[basics('%s')]", name), true
	}

	switch p.Type {
	case "securestring":
		element["type"] = "Microsoft.Common.PasswordBox"
		element["constraints"] = map[string]any{"required": true}
		element["options"] = map[string]any{"hideConfirmation": true}
	case "bool":
		element["type"] = "Microsoft.Common.CheckBox"
		if def, ok := p.DefaultValue.(bool); ok && def {
			element["defaultValue"] = true
		}
	case "int":
		element["type"] = "Microsoft.Common.TextBox"
		element["constraints"] = map[string]any{
			"required":          true,
			"regex":             "^-?[0-9]+$",
			"validationMessage": "Enter a whole number.",
		}
		if p.DefaultValue != nil {
			element["defaultValue"] = fmt.Sprintf("%v", p.DefaultValue)
		}
		return element, fmt.Sprintf("[int(basics('%s'))]", name), true
	case "string":
		element["type"] = "Microsoft.Common.TextBox"
		def, hasDefault := p.DefaultValue.(string)
		element["constraints"] = map[string]any{"required": !hasDefault || def != ""}
		if hasDefault {
			element["defaultValue"] = def
		}
	default:
		return nil, "", false
	}

	return element, fmt.Sprintf("[basics('%s')]", name), true
}

func humanizeParamName(name string) string {
	if name == "" {
		return ""
	}

	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(runes[i-1]) {
			b.WriteRune(' ')
		}
		if i == 0 {
			r = unicode.ToUpper(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}
