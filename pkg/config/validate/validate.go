package validate

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/config"
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
			return ValidateVars(ctx, a)
		},
		func() error {
			return ValidatePolicies(a)
		},
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
