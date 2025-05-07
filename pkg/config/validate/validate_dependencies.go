package validate

import (
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func ValidateDependencies(cfg *config.AppConfig) error {
	componentNames := make([]string, 0)
	for _, v := range cfg.Components {
		componentNames = append(componentNames, v.Name)
	}

	for _, comp := range cfg.Components {
		if generics.SliceContains(comp.Name, comp.Dependencies) {
			return errs.NewUserFacing("Validation error: component depends on itself (circular dependency)")
		}

		for _, dep := range comp.Dependencies {
			if !generics.SliceContains(dep, componentNames) {
				return errs.NewUserFacing("Validation error: component dependency does not exist (%s)", dep)
			}
		}

		if len(generics.UniqueSlice(comp.Dependencies)) != len(comp.Dependencies) {
			return errs.NewUserFacing("Validation error: one or more dependencies were duplicated.")
		}
	}

	return nil
}
