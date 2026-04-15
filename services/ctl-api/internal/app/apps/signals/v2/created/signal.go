package created

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "app-created"

type Signal struct {
	signal.Hooks
	AppID string `json:"app_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppID == "" {
		return errors.New("app_id is required")
	}

	// Validate app exists
	_, err := activities.AwaitGetByAppID(ctx, s.AppID)
	if err != nil {
		return errors.Wrap(err, "app not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	// Get the app
	currentApp, err := activities.AwaitGetByAppID(ctx, s.AppID)
	if err != nil {
		return errors.Wrap(err, "unable to get app from database")
	}

	// Send notification
	s.sendNotification(ctx, l, notifications.NotificationsTypeAppCreated, s.AppID, map[string]string{
		"app_name":   currentApp.Name,
		"created_by": currentApp.CreatedBy.Email,
	})

	// TODO: Add analytics tracking
	// Original code: w.analytics.Track(ctx, events.AppCreated, map[string]any{"app_id": appID})
	// This requires analytics writer which is not available in signal context.
	// We'll need to either:
	// 1. Add analytics as an activity call, or
	// 2. Pass analytics writer through signal struct

	return nil
}

// sendNotification sends email and slack notifications (inlined from worker/notification.go)
func (s *Signal) sendNotification(ctx workflow.Context, l *zap.Logger, typ notifications.Type, appID string, vars map[string]string) {
	// Send email
	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send email",
			zap.Error(err),
			zap.String("type", typ.String()))
	}

	// Send slack notification
	if err := sharedactivities.AwaitSendSlack(ctx, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send slack notification",
			zap.Error(err),
			zap.String("type", typ.String()))
	}
}
