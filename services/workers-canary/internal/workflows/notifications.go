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
	// provision
	notificationTypeCanaryStart notificationType = iota + 1
	notificationTypeProvisionError
	notificationTypeTestsError
	notificationTypeDeprovisionError
	notificationTypeCanarySuccess
)

func (n notificationType) notification(canaryID, env string, sandboxMode bool, err error) string {
	idSnippet := fmt.Sprintf("`%s` canary `%s`", env, canaryID)
	if sandboxMode {
		idSnippet = fmt.Sprintf("`%s` sandbox canary `%s`", env, canaryID)
	}

	switch n {

	// provision notifs
	case notificationTypeCanaryStart:
		return fmt.Sprintf("🐦 started %s 🚂", idSnippet)
	case notificationTypeCanarySuccess:
		return fmt.Sprintf("🐦 finished %s 🏁", idSnippet)
	case notificationTypeProvisionError:
		return fmt.Sprintf("🐦 error provisioning %s\n\t```%s```", idSnippet, err)
	case notificationTypeTestsError:
		return fmt.Sprintf("🐦 error running test script %s\n\t```%s```", idSnippet, err)
	case notificationTypeDeprovisionError:
		return fmt.Sprintf("🐦 error deprovisioning %s\n\t```%s```", idSnippet, err)
	}

	return ""
}

func (w *wkflow) sendNotification(ctx workflow.Context, typ notificationType, canaryID string, sandboxMode bool, stepErr error) {
	msg := typ.notification(canaryID, w.cfg.Env.String(), sandboxMode, stepErr)
	l := workflow.GetLogger(ctx)
	if w.cfg.DisableNotifications {
		l.Info(msg)
		return
	}

	if err := sharedactivities.SendNotification(ctx, &sharedactivitiesv1.SendNotificationRequest{
		SlackWebhookUrl: w.cfg.SlackWebhookURL,
		Notification:    msg,
	}); err != nil {
		l.Error("failed to send notification", zap.Error(err))
	}
}
