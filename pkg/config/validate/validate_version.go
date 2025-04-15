package validate

import (
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	currentVersion string = "v1"
)

func ValidateVersion(a *config.AppConfig) error {
	if !generics.SliceContains(a.Version, []string{currentVersion, "v2"}) {
		return config.ErrConfig{
			Description: "version must be v1 or v2",
		}
	}
	return nil
}
