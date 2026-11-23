package build

import (
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

func RegistryRetentionConfig(retention *config.RegistryRetention, appID, appConfigID string) (*app.AppRegistryRetentionConfig, error) {
	if retention == nil {
		return nil, errors.New("registry retention config is required")
	}
	if err := validation.ValidateRegistryRetentionMode(retention.Mode); err != nil {
		return nil, err
	}
	if err := validation.ValidateRegistryRetentionRetainCount(retention.RetainCount); err != nil {
		return nil, err
	}

	return &app.AppRegistryRetentionConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Mode:        retention.Mode,
		RetainCount: retention.RetainCount,
	}, nil
}
