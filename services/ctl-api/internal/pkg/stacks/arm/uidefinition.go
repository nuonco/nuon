package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// QuickLinkUIDefinition renders the createUiDefinition that accompanies the quick
// link's wrapper template.
//
// Without it the portal renders its stock Basics step, where subscription,
// resource group, and location are free choices that carry nothing over between
// visits. That is fine for a first deploy and wrong for every one after it: the
// wrapper always names the stack <install-id>-stack, but a stack is identified by
// scope *and* name, so a customer who picks a different resource group on a
// reprovision creates a second, independent stack rather than updating the
// install's. Nothing errors — the install simply stops converging, and the
// duplicate keeps its own deny assignments.
//
// The UI definition closes that by constraining the step to the values the
// install already committed to:
//
//   - resourceGroup rejects any name but the install's, with allowExisting
//     because by definition it already holds resources on every deploy after the
//     first.
//   - location is pinned to the install's region. Deploying elsewhere would strand
//     resources in a region the platform does not track.
//   - subscription requires deploymentStacks/write, so a missing permission shows
//     up in the form rather than as a mid-deploy authorization failure.
//
// Parameters carrying defaults are left out of outputs, so the wrapper's defaults
// apply. Parameters without one — customer secrets, which render as securestring
// — get a field on the Basics step, since there is nowhere else for their value
// to come from.
func (t *Templates) QuickLinkUIDefinition(inp *stacks.TemplateInput) ([]byte, string, error) {
	wrapperParams, err := t.quickLinkWrapperParameters(inp)
	if err != nil {
		return nil, "", err
	}

	scope := scopeFor(inp)
	location := inp.Install.AzureAccount.Location

	// A deployment stack is identified by scope AND name, so pinning the name is
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
			"constraints": map[string]any{"validations": subscriptionValidations},
		},
		"location": map[string]any{
			"allowedValues": []string{location},
			"toolTip":       "The install's region. It is fixed for the lifetime of the install.",
		},
	}

	// At subscription scope the stack template creates the install resource group
	// itself, so the portal shows no resource group picker to constrain.
	if !scope.subscription {
		// The customer names the group on the first deploy — any name is fine, and
		// the group is theirs to choose. Every deploy after that has to land in the
		// same one, or it creates a second stack rather than updating this install.
		// Which case we are in is told by whether the stack has phoned home its
		// resource group yet.
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

	for _, name := range sortedParamNames(wrapperParams) {
		p := wrapperParams[name]
		if p.DefaultValue != nil || name == "location" {
			continue
		}

		element := map[string]any{
			"name":    name,
			"label":   name,
			"toolTip": "",
		}
		if p.Metadata != nil && p.Metadata.Description != "" {
			element["toolTip"] = p.Metadata.Description
		}

		if p.Type == "securestring" {
			element["type"] = "Microsoft.Common.PasswordBox"
			element["constraints"] = map[string]any{"required": true}
			element["options"] = map[string]any{"hideConfirmation": true}
			basics = append(basics, element)
			outputs[name] = fmt.Sprintf("[basics('%s')]", name)
			continue
		}

		element["type"] = "Microsoft.Common.TextBox"
		element["constraints"] = map[string]any{"required": true}
		basics = append(basics, element)
		outputs[name] = fmt.Sprintf("[basics('%s')]", name)
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

// deployedResourceGroupName is the resource group the install's stack actually
// landed in, as reported by the phone-home script. Empty until the first deploy
// completes, which is exactly the window in which the customer is still free to
// name the group whatever they like.
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
