package workflows

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	sharedactivitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
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

func (n notificationType) notification(canaryID, env string, sandboxMode bool, err error, attrs map[string]string) string {
	idSnippet := fmt.Sprintf("`%s` canary `%s`", env, canaryID)
	if sandboxMode {
		idSnippet = fmt.Sprintf("`%s` sandbox canary `%s`", env, canaryID)
	}

	var msg string
	switch n {
	// provision notifs
	case notificationTypeCanaryStart:
		msg = fmt.Sprintf("🐦 started %s 🚂", idSnippet)
	case notificationTypeCanarySuccess:
		msg = fmt.Sprintf("🐦 finished %s 🏁", idSnippet)
	case notificationTypeProvisionError:
		msg = fmt.Sprintf("🐦 error provisioning %s\n\t```%s```", idSnippet, err)
	case notificationTypeTestsError:
		msg = fmt.Sprintf("🐦 error running test script %s\n\t```%s```", idSnippet, err)
	case notificationTypeDeprovisionError:
		msg = fmt.Sprintf("🐦 error deprovisioning %s\n\t```%s```", idSnippet, err)
	default:
	}

	for k, v := range attrs {
		msg = fmt.Sprintf("%s\n\t- *%s* - `%s`", msg, k, v)
	}

	return msg
}

func (w *wkflow) sendNotification(ctx workflow.Context, typ notificationType, canaryID string, sandboxMode bool, stepErr error, attrs map[string]string) {
	msg := typ.notification(canaryID, w.cfg.Env.String(), sandboxMode, stepErr, attrs)
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
