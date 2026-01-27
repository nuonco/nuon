package syncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// syncApp updates the app metadata (description, display name, slack webhook).
func (s *syncer) syncApp(ctx context.Context) error {
	currentApp := app.App{
		ID: s.appID,
	}

	updates := app.App{
		Description: generics.NewNullString(s.cfg.Description),
		DisplayName: generics.NewNullString(s.cfg.DisplayName),
	}

	res := s.db.WithContext(ctx).
		Model(&currentApp).
		Updates(updates)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to update app",
			Err:         fmt.Errorf("unable to update app: %w", res.Error),
		}
	}

	// Update slack webhook URL in notifications config if provided
	if s.cfg.SlackWebhookURL != "" {
		res = s.db.WithContext(ctx).
			Select("slack_webhook_url").
			Model(&app.NotificationsConfig{}).
			Where(&app.NotificationsConfig{
				OwnerID: currentApp.ID,
			}).
			Updates(app.NotificationsConfig{
				SlackWebhookURL: s.cfg.SlackWebhookURL,
			})
		if res.Error != nil {
			return sync.SyncInternalErr{
				Description: "unable to sync app notifications config",
				Err:         fmt.Errorf("unable to sync app notifications config: %w", res.Error),
			}
		}
	}

	return nil
}
