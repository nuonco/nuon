package validate

import (
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/errs"
)

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
