package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SendNotificationRequest struct {
	OrgID string
	AppID string

	Type notifications.Type
	Vars map[string]string
}

func (a *Activities) SendNotification(ctx context.Context, req SendNotificationRequest) error {
	cfg, err := a.getNotificationsConfig(ctx, req.OrgID, req.AppID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	err = a.notifs.Send(ctx, cfg, req.Type, req.Vars)
	if err != nil {
		return fmt.Errorf("unable to send notification: %w", err)
	}

	return nil
}

func (a *Activities) getNotificationsConfig(ctx context.Context, orgID, appID string) (*app.NotificationsConfig, error) {
	ownerType := "orgs"
	ownerID := orgID
	if appID != "" {
		ownerType = "apps"
		ownerID = appID
	}

	notifCfg := app.NotificationsConfig{}
	res := a.db.WithContext(ctx).
		Where(app.NotificationsConfig{
			OwnerType: ownerType,
			OwnerID:   ownerID,
		}).
		First(&notifCfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get notifications config: %w", res.Error)
	}

	return &notifCfg, nil
}
