package validate

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/config"
)

func Validate(ctx context.Context, v *validator.Validate, a *config.AppConfig) error {
	fns := []func() error{
		func() error {
			return ValidateVersion(a)
		},
		func() error {
			return v.Struct(a)
		},
		func() error {
			return ValidateJSONSchema(ctx, a)
		},
		func() error {
			return ValidateDuplicateComponentNames(a)
		},
		func() error {
			return ValidateDependencies(a)
		},
		func() error {
			return ValidateActionWorkflowTriggers(a)
		},
		func() error {
			return ValidatePolicies(a)
		},
		func() error {
			return ValidateTriggers(a)
		},

		func() error {
			if a.Secrets != nil {
				return a.Secrets.Validate()
			}
			return nil
		},
		func() error {
			if a.Components != nil {
				return a.Components.Validate()
			}
			return nil
		},
		func() error {
			for _, install := range a.Installs {
				if err := install.Validate(); err != nil {
					return err
				}
			}
			return nil
		},
		func() error {
			return ValidateDefaultLabels(a)
		},
		func() error {
			return nil
		},

		func() error {
			return a.Permissions.Validate()
		},
		func() error {
			return a.BreakGlass.Validate()
		},

		func() error {
			return ValidateTemplateRefs(a)
		},
		func() error {
			return ValidateCustomNestedStackOutputs(a)
		},
		func() error {
			return ValidateAzureRunnerIdentities(a)
		},
		func() error {
			return validateAzureCustomNestedStacks(a)
		},
		func() error {
			if err := a.OperationRoles.Validate(); err != nil {
				return err
			}
			if err := a.OperationRoles.ValidateWithConfig(
				a.Components,
				a.Actions,
				a.Permissions,
				a.BreakGlass); err != nil {
				return err
			}
			return nil
		},
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func validateAzureCustomNestedStacks(a *config.AppConfig) error {
	if a.Stack == nil {
		return nil
	}
	if err := config.ValidateAzureCustomNestedStacks(a.Stack.Type, a.Stack.CustomNestedStacks); err != nil {
		return err
	}
	for _, install := range a.Installs {
		if install.StackOverrides == nil {
			continue
		}
		if err := config.ValidateAzureCustomNestedStacks(a.Stack.Type, install.StackOverrides.CustomNestedStacks); err != nil {
			return err
		}
	}
	return nil
}
