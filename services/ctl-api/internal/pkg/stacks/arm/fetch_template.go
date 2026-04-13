package arm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	fetchTimeout     = 10 * time.Second
	maxTemplateBytes = 10 << 20 // 10 MB
)

type armTemplateResource struct {
	Type           string                 `json:"type"`
	Name           string                 `json:"name,omitempty"`
	DependsOn      json.RawMessage        `json:"dependsOn,omitempty"`
	SubscriptionId string                 `json:"subscriptionId,omitempty"`
	Properties     *armResourceProperties `json:"properties,omitempty"`
}

type armResourceProperties struct {
	Template *struct {
		Resources []armTemplateResource `json:"resources,omitempty"`
	} `json:"template,omitempty"`
}

type armTemplateShape struct {
	Parameters map[string]struct {
		Type         string `json:"type"`
		DefaultValue any    `json:"defaultValue,omitempty"`
		Metadata     *struct {
			Description string `json:"description,omitempty"`
		} `json:"metadata,omitempty"`
	} `json:"parameters"`
	Resources []armTemplateResource `json:"resources"`
	Outputs   map[string]struct{}   `json:"outputs"`
}

// hasManagedIdentity returns true if the template declares a
// Microsoft.ManagedIdentity/userAssignedIdentities resource.
func (t *armTemplateShape) hasManagedIdentity() bool {
	for _, r := range t.Resources {
		if r.Type == "Microsoft.ManagedIdentity/userAssignedIdentities" {
			return true
		}
	}
	return false
}

func fetchARMTemplate(templateURL string) (*armTemplateShape, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(templateURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template %s: %w", templateURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, templateURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTemplateBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read template body: %w", err)
	}

	var tmpl armTemplateShape
	if err := json.Unmarshal(body, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse ARM template: %w", err)
	}

	return &tmpl, nil
}

// extractARMParameters returns two maps:
// 1. params: all parameter names from the template (including reserved)
// 2. hoistedParams: parameter definitions to hoist into the parent template (excluding reserved)
func extractARMParameters(tmpl *armTemplateShape, reservedNames []string) (map[string]bool, map[string]ARMParameter) {
	params := map[string]bool{}
	hoistedParams := map[string]ARMParameter{}

	for paramName, param := range tmpl.Parameters {
		params[paramName] = true

		if slices.Contains(reservedNames, paramName) {
			continue
		}

		hp := ARMParameter{
			Type:         param.Type,
			DefaultValue: param.DefaultValue,
		}
		if param.Metadata != nil && param.Metadata.Description != "" {
			hp.Metadata = &ARMParameterMetadata{
				Description: param.Metadata.Description,
			}
		}
		hoistedParams[paramName] = hp
	}

	return params, hoistedParams
}

// validateARMTemplate performs structural validation on a custom nested stack
// ARM template. It catches issues that would only surface at deploy time, such
// as subscription-level nested deployments (which ARM does not support inside
// linked deployments) and invalid dependsOn references.
func validateARMTemplate(tmpl *armTemplateShape) error {
	var errs []string

	// Build set of resource names declared at the top level for dependsOn validation.
	resourceNames := map[string]bool{}
	for _, r := range tmpl.Resources {
		if r.Name != "" {
			resourceNames[r.Name] = true
		}
	}

	for i, r := range tmpl.Resources {
		resourceLabel := r.Name
		if resourceLabel == "" {
			resourceLabel = fmt.Sprintf("index %d", i)
		}

		// Subscription-level nested deployments are not supported inside linked
		// deployments. ARM silently scopes them to resource-group level causing
		// confusing "resource is not defined in the template" errors.
		if r.Type == "Microsoft.Resources/deployments" && r.SubscriptionId != "" {
			errs = append(errs, fmt.Sprintf(
				"resource %q: subscription-level nested deployments (subscriptionId set) "+
					"are not supported inside linked deployments; move the role assignment "+
					"to the parent template or use a managed identity output pattern",
				resourceLabel,
			))
		}

		// Validate dependsOn references point to resources declared in this template.
		deps, err := parseDependsOn(r.DependsOn)
		if err != nil {
			errs = append(errs, fmt.Sprintf("resource %q: invalid dependsOn: %v", resourceLabel, err))
			continue
		}
		for _, dep := range deps {
			// ARM dependsOn can use either plain resource names or full
			// resourceId expressions like "[resourceId(...)]". We can only
			// validate plain name references.
			if strings.HasPrefix(dep, "[") {
				continue
			}
			if !resourceNames[dep] {
				errs = append(errs, fmt.Sprintf(
					"resource %q: dependsOn references %q which is not defined in this template",
					resourceLabel, dep,
				))
			}
		}

		// Check for subscription-level deployments nested inside inline
		// template blocks (e.g., deployment resources with inner templates
		// that themselves contain subscription-scoped deployments).
		if r.Type == "Microsoft.Resources/deployments" && r.Properties != nil &&
			r.Properties.Template != nil {
			for j, nested := range r.Properties.Template.Resources {
				if nested.Type == "Microsoft.Resources/deployments" && nested.SubscriptionId != "" {
					nestedLabel := nested.Name
					if nestedLabel == "" {
						nestedLabel = fmt.Sprintf("index %d", j)
					}
					errs = append(errs, fmt.Sprintf(
						"resource %q: inline template contains subscription-level nested deployment %q; "+
							"this is not supported inside linked deployments",
						resourceLabel, nestedLabel,
					))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("ARM template validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// parseDependsOn extracts the dependency list from a raw JSON value.
// ARM templates allow dependsOn to be either a JSON array of strings or a
// single string.
func parseDependsOn(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// Try array first (most common).
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}

	// Try single string.
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}

	return nil, fmt.Errorf("expected string or array of strings, got: %s", string(raw))
}
