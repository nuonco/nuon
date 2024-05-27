package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration045NotificationsConfigExistingOrgs(ctx context.Context) error {
	var orgs []app.Org

	res := m.db.WithContext(ctx).
		Preload("NotificationsConfig").
		Preload("CreatedBy").
		Find(&orgs)
	if res.Error != nil {
		return fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	for _, org := range orgs {
		m.l.Info("processing org", zap.String("name", org.Name))

		if org.NotificationsConfigID != "" {
			m.l.Info("org already has notifications config")
			continue
		}

		m.l.Info("creating org notifications config")
		notificationsCfg := app.NotificationsConfig{
			CreatedByID:              org.CreatedByID,
			OrgID:                    org.ID,
			EnableSlackNotifications: org.CreatedBy.TokenType == app.TokenTypeAuth0,
			EnableEmailNotifications: org.CreatedBy.TokenType == app.TokenTypeAuth0,
			InternalSlackWebhookURL:  m.cfg.InternalSlackWebhookURL,
			OwnerID:                  org.ID,
			OwnerType:                "orgs",
		}

		// create notifications config
		res := m.db.WithContext(ctx).
			Create(&notificationsCfg)
		if res.Error != nil {
			return fmt.Errorf("unable to create notifications config: %w", res.Error)
		}

                // update org
		res = m.db.WithContext(ctx).Model(&app.Org{
			ID: org.ID,
		}).Updates(app.Org{
			NotificationsConfigID: notificationsCfg.ID,
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update org with notifications config: %w", res.Error)
		}
	}
	return nil
}
