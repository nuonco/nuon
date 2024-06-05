package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type SyncAppMetadataRequest struct {
	AppConfigID string
}

func (a *Activities) SyncAppMetadata(ctx context.Context, req SyncAppMetadataRequest) error {
	appCfg, err := a.GetAppConfig(ctx, GetAppConfigRequest{req.AppConfigID})
	if err != nil {
		return err
	}

	cfg, err := parse.Parse(parse.ParseConfig{
		Bytes:       []byte(appCfg.Content),
		BackendType: config.BackendTypeS3,
		Template:    true,
		V:           validator.New(),
		Context:     config.ConfigContextConfigOnly,
	})
	if err != nil {
		return fmt.Errorf("unable to parse config file: %w", err)
	}

	res := a.db.WithContext(ctx).
		Model(&app.App{
			ID: appCfg.AppID,
		}).
		Updates(app.App{
			Description: cfg.Description,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to sync app metadata: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("app not found %w", gorm.ErrRecordNotFound)
	}

	res = a.db.WithContext(ctx).
		Model(&app.NotificationsConfig{}).
		Where(app.NotificationsConfig{
			OwnerID: appCfg.AppID,
		}).
		Updates(app.NotificationsConfig{
			SlackWebhookURL: cfg.SlackWebhookURL,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to sync app notifications config: %w", res.Error)
	}

	return nil
}
