package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// @temporal-gen workflow
func (w *Workflows) Created(ctx workflow.Context, sreq signals.RequestSignal) error {
	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeOrgCreated, sreq.ID, map[string]string{
		"org_name":   org.Name,
		"created_by": org.CreatedBy.Email,
		"email":      org.CreatedBy.Email,
		"org_url":    fmt.Sprintf("%s/%s", w.cfg.AppURL, org.ID),
	})

	w.analytics.Track(ctx, events.OrgCreated, map[string]interface{}{
		"org_id":   org.ID,
		"org_type": org.OrgType,
	})

	return nil
}
