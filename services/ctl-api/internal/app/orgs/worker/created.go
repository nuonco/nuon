package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

func (w *Workflows) created(ctx workflow.Context, orgID string, _ bool) error {
	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		w.updateStatus(ctx, orgID, StatusError, "unable to get org from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeOrgCreated, orgID, map[string]string{
		"org_name":   org.Name,
		"created_by": org.CreatedBy.Email,
		"email":      org.CreatedBy.Email,
	})
	return nil
}
