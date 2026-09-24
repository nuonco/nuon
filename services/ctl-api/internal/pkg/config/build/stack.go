package build

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func StackConfig(stack *config.StackConfig, appID, appConfigID string) (*app.AppStackConfig, error) {
	if stack == nil {
		return nil, errors.New("stack config is required")
	}

	stackType := app.StackType(stack.Type)

	if err := validateNestedTemplateURL(stackType, stack.VPCNestedTemplateURL, "vpc_nested_template_url"); err != nil {
		return nil, err
	}
	if err := validateNestedTemplateURL(stackType, stack.RunnerNestedTemplateURL, "runner_nested_template_url"); err != nil {
		return nil, err
	}
	if err := config.ValidateDeploymentScope(stack.DeploymentScope, stack.Type); err != nil {
		return nil, err
	}
	if err := config.ValidateAzureCustomNestedStacks(stack.Type, stack.CustomNestedStacks); err != nil {
		return nil, err
	}
	if err := config.ValidateGCPCustomNestedStacks(stack.Type, stack.CustomNestedStacks); err != nil {
		return nil, err
	}

	// Copy so marking the upload status pending does not mutate the caller's
	// parsed config.
	customNestedStacks := make([]config.CustomNestedStack, 0, len(stack.CustomNestedStacks))
	seenNames := make(map[string]int, len(stack.CustomNestedStacks))
	seenIndices := make(map[int]string, len(stack.CustomNestedStacks))
	for i, nested := range stack.CustomNestedStacks {
		if nested.Name == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if prev, exists := seenNames[nested.Name]; exists {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): name is already used by custom_nested_stacks[%d]; each stack must have a unique name", i, nested.Name, prev)
		}
		seenNames[nested.Name] = i
		if prev, exists := seenIndices[nested.Index]; exists {
			return nil, fmt.Errorf("custom_nested_stacks: index %d is used by both %q and %q; each stack must have a unique index", nested.Index, prev, nested.Name)
		}
		seenIndices[nested.Index] = nested.Name
		if nested.TemplateURL == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, nested.Name)
		}
		if stackType != app.StackTypeGCP && nested.Contents == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): contents is required when template_url is set", i, nested.Name)
		}
		for paramName, paramValue := range nested.Parameters {
			if err := config.ValidateStackParameterTemplate(paramValue); err != nil {
				return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q: %w", i, nested.Name, paramName, err)
			}
		}

		if stackType == app.StackTypeGCP {
			nested.Status = config.CustomNestedStackStatusReady
		} else {
			nested.Status = config.CustomNestedStackStatusPending
		}
		customNestedStacks = append(customNestedStacks, nested)
	}

	return &app.AppStackConfig{
		AppID:                   appID,
		AppConfigID:             appConfigID,
		Type:                    stackType,
		Name:                    stack.Name,
		Description:             stack.Description,
		VPCNestedTemplateURL:    stack.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: stack.RunnerNestedTemplateURL,
		// Not normalized: empty means resource group, and rewriting it would make
		// every pre-existing config diff on its next sync.
		DeploymentScope:    app.StackDeploymentScope(stack.DeploymentScope),
		CustomNestedStacks: customNestedStacks,
	}, nil
}

func validateNestedTemplateURL(stackType app.StackType, url, field string) error {
	if url == "" {
		return nil
	}
	if stackType == app.StackTypeAzure {
		return config.ValidateHTTPSURL(url, field)
	}
	return config.ValidateTemplateURL(url, field)
}
