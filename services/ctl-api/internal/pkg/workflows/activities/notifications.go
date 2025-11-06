package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

type SendNotificationRequest struct {
	OrgID string `validate:"required"`
	AppID string `validate:"required"`

	Type notifications.Type `validate:"required"`
	Vars map[string]string  `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 10s
func (a *Activities) SendEmail(ctx context.Context, req SendNotificationRequest) error {
	cfg, err := a.getNotificationsConfig(ctx, req.OrgID, req.AppID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	err = a.notifs.SendEmail(ctx, cfg, req.Type, req.Vars)
	if err != nil {
		return fmt.Errorf("unable to send email notification: %w", err)
	}

	return nil
}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 10s
func (a *Activities) SendSlack(ctx context.Context, req SendNotificationRequest) error {
	cfg, err := a.getNotificationsConfig(ctx, req.OrgID, req.AppID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	err = a.notifs.SendSlack(ctx, cfg, req.Type, req.Vars)
	if err != nil {
		return fmt.Errorf("unable to send slack notification: %w", err)
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
