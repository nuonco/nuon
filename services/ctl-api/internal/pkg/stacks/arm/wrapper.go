package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	deploymentStacksAPIVersion = "2025-07-01"
)

func installStackName(installID string) string {
	return fmt.Sprintf("%s-stack", installID)
}

// why: quickLinkWrapperParameters is the parameter set the wrapper re-declares from the
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
			// why: resourcesWithoutDeleteSupport defaults to "fail" server-side; it is
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

	// why: location is required at subscription scope and rejected at resource-group
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
