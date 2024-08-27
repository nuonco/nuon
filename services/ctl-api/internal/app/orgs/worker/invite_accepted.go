package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
)

// inviteAccepted: is called when an invite is accepted
func (w *Workflows) inviteAccepted(ctx workflow.Context, orgID, inviteID string) error {
	org, err := activities.AwaitGetByOrgID(ctx, orgID)
	if err != nil {
		w.updateStatus(ctx, orgID, app.OrgStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	orgInvite, err := activities.AwaitGetInviteByInviteID(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("unable to get org invite: %w", err)
	}

	w.sendNotification(ctx, notifications.NotificationsTypeOrgInviteAccepted, orgID, map[string]string{
		"email":    orgInvite.Email,
		"org_name": org.Name,
	})
	return nil
}
