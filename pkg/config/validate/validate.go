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
			return ValidatePolicies(a)
		},

		// NOTE(jm): we are moving validation functions for types into the actual types.
		// We build this validation tooling here, so we can validate as many things up front as possible.
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

		// TBH, this does not really work
		func() error {
			return nil
			return ValidateVars(ctx, a)
		},
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
