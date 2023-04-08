package workflows

import (
	"fmt"

	sharedactivitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type notificationType int

const (
	notificationTypeProvisionStart notificationType = iota + 1
	notificationTypeProvisionSuccess
	notificationTypeProvisionError

	notificationTypeDeprovisionStart
	notificationTypeDeprovisionSuccess
	notificationTypeDeprovisionError
)

func (n notificationType) notification(canaryID string, err error) string {
	switch n {
	case notificationTypeProvisionStart:
		return fmt.Sprintf("🐦 started provisioning canary `%s` 🚂", canaryID)
	case notificationTypeProvisionSuccess:
		return fmt.Sprintf("🐦 successfully provisioning canary `%s` 🏁", canaryID)
	case notificationTypeProvisionError:
		return fmt.Sprintf("🐦 error provisioning canary `%s`\n\t```%s```", canaryID, err)
	case notificationTypeDeprovisionStart:
		return fmt.Sprintf("🐦 started deprovisioning canary `%s` 👷", canaryID)
	case notificationTypeDeprovisionSuccess:
		return fmt.Sprintf("🐦 successfully deprovisioned canary `%s` 🏁", canaryID)
	case notificationTypeDeprovisionError:
		return fmt.Sprintf("🐦 error deprovisioning canary `%s`\n\t```%s```", canaryID, err)
	}

	return ""
}

func (w *wkflow) sendNotification(ctx workflow.Context, typ notificationType, canaryID string, stepErr error) {
	msg := typ.notification(canaryID, stepErr)
	l := zap.L()

	if err := sharedactivities.SendNotification(ctx, &sharedactivitiesv1.SendNotificationRequest{
		SlackWebhookUrl: w.cfg.SlackWebhookURL,
		Notification:    msg,
	}); err != nil {
		l.Error("failed to send notification", zap.Error(err))
	}
}
