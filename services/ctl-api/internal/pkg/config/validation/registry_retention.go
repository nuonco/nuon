package validation

import (
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func ValidateRegistryRetentionMode(mode *string) error {
	if err := config.ValidateRegistryRetentionMode(mode); err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_registry_retention_mode"),
			Code:        "invalid_registry_retention_mode",
			Description: err.Error(),
		}
	}
	return nil
}

func ValidateRegistryRetentionRetainCount(count *int) error {
	if err := config.ValidateRegistryRetentionRetainCount(count); err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_registry_retention_retain_count"),
			Code:        "invalid_registry_retention_retain_count",
			Description: err.Error(),
		}
	}
	return nil
}

func ValidateRegistryRetentionGracePeriodDays(days *int) error {
	if err := config.ValidateRegistryRetentionGracePeriodDays(days); err != nil {
		return stderr.ErrUser{
			Err:         errors.New("invalid_registry_retention_grace_period_days"),
			Code:        "invalid_registry_retention_grace_period_days",
			Description: err.Error(),
		}
	}
	return nil
}
