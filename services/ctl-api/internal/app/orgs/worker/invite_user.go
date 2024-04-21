package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

// inviteUser invites a user to the org using an email
func (w *Workflows) inviteUser(ctx workflow.Context, orgID, email string) error {
	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: orgID,
	}, &org); err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	if org.CreatedBy.TokenType == app.TokenTypeCanary || org.CreatedBy.TokenType == app.TokenTypeIntegration {
		return nil
	}

	if err := w.defaultExecErrorActivity(ctx, w.acts.SendEmail, activities.SendEmailRequest{
		TransactionalEmailID: loops.OrgInviteEmailID,
		Email:                email,
		Variables: map[string]string{
			"org_name": org.Name,
		},
	}); err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}

	return nil
}
