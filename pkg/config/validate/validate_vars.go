package validate

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/vars"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func ValidateVars(ctx context.Context, cfg *config.AppConfig) error {
	if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
		Vars:                 generics.MapValuesToSlice(cfg.Sandbox.VarsMap),
		Cfg:                  cfg,
		IgnoreSandboxOutputs: true,
	}); err != nil {
		return config.ErrConfig{
			Description: "unable to validate sandbox vars",
			Warning:     true,
			Err:         err,
		}
	}
	if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
		Vars:                 generics.MapValuesToSlice(cfg.Sandbox.EnvVarMap),
		Cfg:                  cfg,
		IgnoreSandboxOutputs: true,
	}); err != nil {
		return config.ErrConfig{
			Description: "unable to validate sandbox vars",
			Warning:     true,
			Err:         err,
		}
	}

	for _, comp := range cfg.Components {
		ui.Step(ctx, "validating vars for component"+comp.Name)
		if err := vars.ValidateVars(ctx, vars.ValidateVarsParams{
			Vars:                 comp.AllVars(),
			Cfg:                  cfg,
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
