package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

// inviteCreated: is called when a new org invite is created_by
func (w *Workflows) inviteUser(ctx workflow.Context, orgID, inviteID string) error {
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

	w.sendNotification(ctx, notifications.NotificationsTypeOrgInvite, orgID, map[string]string{
		"email":      orgInvite.Email,
		"org_name":   org.Name,
		"created_by": orgInvite.CreatedBy.Email,
	})
	return nil
}
