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

	// Copy so marking the upload status pending does not mutate the caller's
	// parsed config.
	customNestedStacks := make([]config.CustomNestedStack, 0, len(stack.CustomNestedStacks))
	for i, nested := range stack.CustomNestedStacks {
		if nested.Name == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if nested.TemplateURL == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, nested.Name)
		}
		if nested.Contents == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): contents is required when template_url is set", i, nested.Name)
		}
		for paramName, paramValue := range nested.Parameters {
			if err := config.ValidateStackParameterTemplate(paramValue); err != nil {
				return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q: %w", i, nested.Name, paramName, err)
			}
		}

		// Pending until the contents have been uploaded to S3; consumers gate
		// on Status before generating a stack from these templates.
		nested.Status = config.CustomNestedStackStatusPending
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
		CustomNestedStacks:      customNestedStacks,
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
