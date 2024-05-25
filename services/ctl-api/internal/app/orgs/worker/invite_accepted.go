package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

// inviteAccepted: is called when an invite is accepted
func (w *Workflows) inviteAccepted(ctx workflow.Context, orgID, inviteID string) error {
	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	var orgInvite app.OrgInvite
	if err := w.defaultExecGetActivity(ctx, w.acts.GetInvite, activities.GetInviteRequest{
		InviteID: inviteID,
	}, &orgInvite); err != nil {
		return fmt.Errorf("unable to get org invite: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeOrgInviteAccepted, orgID, map[string]string{
		"email":    orgInvite.Email,
		"org_name": org.Name,
	})
	return nil
}
