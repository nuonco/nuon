package validate

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
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
