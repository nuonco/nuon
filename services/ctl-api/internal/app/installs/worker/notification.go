package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/notifications"
)

func (w *Workflows) sendNotification(ctx workflow.Context, typ notifications.Type, appID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)
	if err := w.defaultExecErrorActivity(ctx, w.acts.SendNotification, sharedactivities.SendNotificationRequest{
		AppID: appID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send notification",
			zap.Error(err),
			zap.String("type", typ.String()))
	}
}
