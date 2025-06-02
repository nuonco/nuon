package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// inviteCreated: is called when a new org invite is created_by
//
// @temporal-gen workflow
func (w *Workflows) InviteUser(ctx workflow.Context, sreq signals.RequestSignal) error {
	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		w.updateStatus(ctx, sreq.ID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	orgInvite, err := activities.AwaitGetInviteByInviteID(ctx, sreq.InviteID)
	if err != nil {
		return fmt.Errorf("unable to get org invite: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeOrgInvite, sreq.ID, map[string]string{
		"email":      orgInvite.Email,
		"org_name":   org.Name,
		"created_by": orgInvite.CreatedBy.Email,
		"login_url":  fmt.Sprintf("%s/api/auth/login", w.cfg.AppURL),
	})
	return nil
}
