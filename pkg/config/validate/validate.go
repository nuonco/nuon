package validate

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/schema"
	"github.com/powertoolsdev/mono/pkg/config/vars"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	currentVersion string = "v1"
)

func ValidateVersion(a *config.AppConfig) error {
	if a.Version != currentVersion {
		return config.ErrConfig{
			Description: "version must be v1",
		}
	}
	return nil
}

func ValidateVars(ctx context.Context, obj *config.AppConfig) error {
	if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
		Vars:                 config.TerraformVariables(obj.Sandbox.Vars),
		Cfg:                  obj,
		IgnoreSandboxOutputs: true,
	}); err != nil {
		return config.ErrConfig{
			Description: "unable to validate sandbox vars",
			Warning:     true,
			Err:         err,
		}
	}
	if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
		Vars:                 generics.MapValuesToSlice(obj.Sandbox.VarMap),
		Cfg:                  obj,
		IgnoreSandboxOutputs: true,
	}); err != nil {
		return config.ErrConfig{
			Description: "unable to validate component vars",
			Warning:     true,
			Err:         err,
		}
	}

	for _, comp := range obj.Components {
		if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
			Vars:                 comp.AllVars(),
			Cfg:                  obj,
			IgnoreSandboxOutputs: true,
		}); err != nil {
			return config.ErrConfig{
				Warning:     true,
				Err:         err,
				Description: "unable to validate component vars",
			}
		}
	}

	return nil
}

func ValidateDuplicateComponentNames(cfg *config.AppConfig) error {
	componentNames := make(map[string]bool)
	for _, v := range cfg.Components {
		if _, ok := componentNames[v.Name]; ok {
			return errs.NewUserFacing("Validation error: duplicate component name %q", v.Name)
		}
		componentNames[v.Name] = true
	}
	return nil
}

func Validate(ctx context.Context, v *validator.Validate, a *config.AppConfig) error {
	fns := []func() error{
		func() error {
			return v.Struct(a)
		},
		func() error {
			return ValidateVersion(a)
		},
		func() error {
			return ValidateDuplicateComponentNames(a)
		},
		func() error {
			_, err := schema.Validate(ctx, a)
			return err
		},
		func() error {
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
