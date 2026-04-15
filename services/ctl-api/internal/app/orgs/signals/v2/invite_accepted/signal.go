package invite_accepted

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "org-invite-accepted"

type Signal struct {
	signal.Hooks
	OrgID    string `json:"org_id"`
	InviteID string `json:"invite_id"`
	LoginURL string `json:"login_url"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	if s.InviteID == "" {
		return fmt.Errorf("invite_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	orgInvite, err := activities.AwaitGetInviteByInviteID(ctx, s.InviteID)
	if err != nil {
		return fmt.Errorf("unable to get org invite: %w", err)
	}

	vars := map[string]string{
		"email":     orgInvite.Email,
		"org_name":  org.Name,
		"login_url": s.LoginURL,
	}

	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{OrgID: s.OrgID, Type: notifications.NotificationsTypeOrgInviteAccepted, Vars: vars}); err != nil {
		l.Error("unable to send email", zap.Error(err))
	}
	if err := sharedactivities.AwaitSendSlack(ctx, sharedactivities.SendNotificationRequest{OrgID: s.OrgID, Type: notifications.NotificationsTypeOrgInviteAccepted, Vars: vars}); err != nil {
		l.Error("unable to send slack notification", zap.Error(err))
	}

	return nil
}
