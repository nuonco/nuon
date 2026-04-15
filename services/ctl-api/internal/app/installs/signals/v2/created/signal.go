package created

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "install-created"

type Signal struct {
	signal.Hooks
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithInit = (*Signal)(nil)

func (s *Signal) Init(_ workflow.Context) error {
	s.Hooks.InstallID = &s.InstallID
	s.Hooks.Operation = "install-created"
	return nil
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get the install
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	// Send notification for first install
	if install.InstallNumber == 1 {
		s.sendNotification(ctx, notifications.NotificationsTypeFirstInstallCreated, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})
	}

	// Send notification for subsequent installs
	if install.InstallNumber > 1 {
		s.sendNotification(ctx, notifications.NotificationsTypeInstallCreated, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})

		// TODO: Add analytics tracking for InstallCreated event
		// This was removed during migration because temporalanalytics.Writer is not available
		// in the signal context. We'll need to either:
		// 1. Add analytics as an activity call, or
		// 2. Pass analytics writer through signal struct
		// Original code: w.analytics.Track(ctx, events.InstallCreated, map[string]any{"install_id": install.ID})
	}

	return nil
}

// sendNotification sends email and slack notifications (inlined from worker/notification.go)
func (s *Signal) sendNotification(ctx workflow.Context, typ notifications.Type, appID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)

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
