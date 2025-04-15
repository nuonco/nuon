package validate

import (
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func ValidateActionWorkflowTriggers(cfg *config.AppConfig) error {
	componentNames := make(map[string]bool)
	for _, v := range cfg.Components {
		componentNames[v.Name] = true
	}

	for _, actCfg := range cfg.Actions {
		for _, trigger := range actCfg.Triggers {
			if !generics.SliceContains(trigger.Type, []string{
				"pre-component-deploy",
				"post-component-deploy",
			}) {
				continue
			}

			if trigger.ComponentName == "" {
				return errs.NewUserFacing("Validation error: %s trigger does not have component_name set", trigger.Type)
			}

			// since the component deploy trigger is being used, make sure this references a valid
			// component.
			if _, ok := componentNames[trigger.ComponentName]; !ok {
				return errs.NewUserFacing(
					"Validation error: %s trigger references an invalid component (%s)",
					trigger.Type,
					trigger.ComponentName,
				)
			}
		}
	}

	return nil
}
