package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration058AWSRegionTypes(ctx context.Context) error {
	var appSandboxConfigs []app.AppSandboxConfig
	res := m.db.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Find(&appSandboxConfigs)
	if res.Error != nil {
		return res.Error
	}

	for _, appSandboxCfg := range appSandboxConfigs {
		fmt.Println("processing")
		if appSandboxCfg.CloudPlatform != app.CloudPlatformAWS {
			continue
		}
		fmt.Println("processing is aws")

		res = m.db.WithContext(ctx).
			Model(app.AppSandboxConfig{
				ID: appSandboxCfg.ID,
			}).
			Updates(app.AppSandboxConfig{
				AWSRegionType: generics.NewNullString(app.AWSRegionTypeDefault.String()),
			})
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}
