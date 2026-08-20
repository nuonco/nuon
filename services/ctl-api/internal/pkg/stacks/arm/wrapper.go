package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	// deploymentStacksAPIVersion is the only API version Microsoft.Resources
	// advertises for deploymentStacks.
	deploymentStacksAPIVersion = "2025-07-01"
)

// installStackName is the deployment stack's name, matching the one the
// documented `az stack group create` command uses so the portal path and the
// CLI path converge on the same stack rather than creating two.
func installStackName(installID string) string {
	return fmt.Sprintf("%s-stack", installID)
}

// quickLinkWrapperParameters is the parameter set the wrapper re-declares from the
// stack template. Shared with QuickLinkUIDefinition so the UI definition's outputs
// cannot drift from the parameters the wrapper actually accepts — a parameter in
// one and not the other either goes unfilled or is rejected outright by the portal.
func (t *Templates) quickLinkWrapperParameters(inp *stacks.TemplateInput) (map[string]ARMParameter, error) {
	inner, err := t.getAzureTemplate(inp)
	if err != nil {
		return nil, err
	}
	return inner.Parameters, nil
}

// QuickLinkWrapper renders the template behind an Azure portal quick link:
// https://portal.azure.com/#create/Microsoft.Template/uri/<encoded wrapper URL>
//
// The portal's Custom Deployment blade only ever creates a plain
// Microsoft.Resources/deployments — it has no deployment-stack mode, and Azure
// exposes stacks only via CLI/PowerShell/REST. This wrapper closes that gap: it
// is a deployment whose single resource is a Microsoft.Resources/deploymentStacks
// pointing at the real template. Deploying it therefore produces the same
// protected, self-cleaning stack as `az stack group create`, keeping both
// denySettings (customers cannot delete managed resources) and actionOnUnmanage
// (resources dropped from the template are deleted on the next reprovision
// rather than orphaned).
//
// The stack template is referenced by URL rather than embedded. Inlining it
// under properties.template does not work: ARM evaluates an inline nested
// template's expressions in the *outer* template's context, so the cross-
// deployment references this template is built from — reference('vnetDeployment')
// and friends — fail to resolve with "could not find template resource or
// resource copy with this name". The usual remedy,
// expressionEvaluationOptions: {scope: inner}, is rejected outright:
// "not supported in 'Microsoft.Resources/deploymentStacks' type of resource".
// A linked template gets its own evaluation context and sidesteps both.
//
// The wrapper re-declares the stack template's parameters and passes them
// through. The portal builds its form from the template it was handed, so
// without this the customer would see no fields at all and any parameter
// lacking a default (customer-supplied secrets, which render as securestring)
// would have nowhere to come from.
func (t *Templates) QuickLinkWrapper(inp *stacks.TemplateInput, templateURL string) ([]byte, string, error) {
	if templateURL == "" {
		return nil, "", fmt.Errorf("unable to render quick link wrapper: template URL is empty")
	}

	params, err := t.quickLinkWrapperParameters(inp)
	if err != nil {
		return nil, "", err
	}

	scope := scopeFor(inp)

	passthrough := make(map[string]any, len(params))
	for name := range params {
		passthrough[name] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", name)}
	}

	stack := map[string]any{
		"type":       "Microsoft.Resources/deploymentStacks",
		"apiVersion": deploymentStacksAPIVersion,
		"name":       installStackName(inp.Install.ID),
		"properties": map[string]any{
			// resourcesWithoutDeleteSupport defaults to "fail" server-side; it is
			// set explicitly so a change in that default cannot silently alter
			// teardown behaviour.
			"actionOnUnmanage": map[string]any{
				"resources":                     "delete",
				"resourceGroups":                "detach",
				"managementGroups":              "detach",
				"resourcesWithoutDeleteSupport": "fail",
			},
			"denySettings": map[string]any{
				"mode":               "denyDelete",
				"applyToChildScopes": false,
			},
			"templateLink": map[string]any{
				"uri": templateURL,
			},
			"parameters": passthrough,
		},
	}

	// location is required at subscription scope and rejected at resource-group
	// scope, where ARM fails the deploy with "The 'location' property is not
	// allowed for '<name>' at resource group scope". Note that
	// `az deployment group validate` accepts it either way — this only surfaces
	// on a real deploy.
	if scope.subscription {
		stack["location"] = inp.Install.AzureAccount.Location
	}

	wrapper := &ARMTemplate{
		Schema:         scope.rootSchema(),
		ContentVersion: "1.0.0.0",
		Parameters:     params,
		Resources:      []any{stack},
	}

	wrapperBytes, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal quick link wrapper: %w", err)
	}

	hash := sha256.Sum256(wrapperBytes)
	checksum := hex.EncodeToString(hash[:])

	return wrapperBytes, checksum, nil
}
