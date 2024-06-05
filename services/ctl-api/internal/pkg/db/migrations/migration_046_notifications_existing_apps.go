package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration046NotifcationsExistingApps(ctx context.Context) error {
	var apps []app.App

	res := m.db.WithContext(ctx).
		Preload("NotificationsConfig").
		Preload("CreatedBy").
		Find(&apps)
	if res.Error != nil {
		return fmt.Errorf("unable to get apps: %w", res.Error)
	}

	for _, currentApp := range apps {
		//if currentApp.NotificationsConfigID != "" {
		//m.l.Info("org already has notifications config")
		//continue
		//}

		notificationsCfg := app.NotificationsConfig{
			CreatedByID:              currentApp.CreatedByID,
			OrgID:                    currentApp.ID,
			EnableSlackNotifications: currentApp.CreatedBy.TokenType == app.TokenTypeAuth0,
			EnableEmailNotifications: currentApp.CreatedBy.TokenType == app.TokenTypeAuth0,
			InternalSlackWebhookURL:  m.cfg.InternalSlackWebhookURL,
			OwnerID:                  currentApp.ID,
			OwnerType:                "apps",
		}

		res := m.db.WithContext(ctx).
			Create(&notificationsCfg)
		if res.Error != nil {
			return fmt.Errorf("unable to create notifications config: %w", res.Error)
		}

		// update app
		res = m.db.WithContext(ctx).Model(&app.App{
			ID: currentApp.ID,
		}).Updates(app.Org{
			NotificationsConfigID: notificationsCfg.ID,
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update app with notifications config: %w", res.Error)
		}
	}
	return nil
}
