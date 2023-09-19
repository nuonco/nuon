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
	notificationTypeProvisionError
	notificationTypeCLICommandsError
	notificationTypeIntrospectionAPIError
	notificationTypeSuccess

	notificationTypeDeprovisionStart
	notificationTypeDeprovisionSuccess
	notificationTypeDeprovisionError
)

func (n notificationType) notification(canaryID, env string, err error) string {
	switch n {
	case notificationTypeProvisionStart:
		return fmt.Sprintf("🐦 started provisioning `%s` canary `%s` 🚂", env, canaryID)
	case notificationTypeSuccess:
		return fmt.Sprintf("🐦 successfully provisioned `%s` canary `%s` 🏁", env, canaryID)
	case notificationTypeProvisionError:
		return fmt.Sprintf("🐦 error provisioning `%s` canary `%s`\n\t```%s```", env, canaryID, err)
	case notificationTypeCLICommandsError:
		return fmt.Sprintf("🐦 error running cli commands `%s` canary `%s`\n\t```%s```", env, canaryID, err)
	case notificationTypeIntrospectionAPIError:
		return fmt.Sprintf("🐦 error introspecting api `%s` canary `%s`\n\t```%s```", env, canaryID, err)
	case notificationTypeDeprovisionStart:
		return fmt.Sprintf("🐦 started deprovisioning `%s` canary `%s` 👷", env, canaryID)
	case notificationTypeDeprovisionSuccess:
		return fmt.Sprintf("🐦 successfully deprovisioned `%s` canary `%s` 🏁", env, canaryID)
	case notificationTypeDeprovisionError:
		return fmt.Sprintf("🐦 error deprovisioning `%s` canary `%s`\n\t```%s```", env, canaryID, err)
	}

	return ""
}

func (w *wkflow) sendNotification(ctx workflow.Context, typ notificationType, canaryID string, stepErr error) {
	msg := typ.notification(canaryID, w.cfg.Env.String(), stepErr)
	l := zap.L()

	if err := sharedactivities.SendNotification(ctx, &sharedactivitiesv1.SendNotificationRequest{
		SlackWebhookUrl: w.cfg.SlackWebhookURL,
		Notification:	 msg,
	}); err != nil {
		l.Error("failed to send notification", zap.Error(err))
	}
}
