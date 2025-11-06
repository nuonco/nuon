package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
)

func (w *Workflows) sendNotification(ctx workflow.Context, typ notifications.Type, orgID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)

	// Send email
	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{
		OrgID: orgID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send email",
			zap.Error(err),
			zap.String("type", typ.String()))
	}

	// Send slack notification
	if err := sharedactivities.AwaitSendSlack(ctx, sharedactivities.SendNotificationRequest{
		OrgID: orgID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send slack notification",
			zap.Error(err),
			zap.String("type", typ.String()))
	}
}
