package created

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "org-created"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return errors.New("org_id is required")
	}

	_, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return errors.Wrap(err, "org not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	org, err := activities.AwaitGetByOrgID(ctx, s.OrgID)
	if err != nil {
		return fmt.Errorf("unable to get org: %w", err)
	}

	s.sendNotification(ctx, notifications.NotificationsTypeOrgCreated, s.OrgID, map[string]string{
		"org_name":   org.Name,
		"created_by": org.CreatedBy.Email,
		"email":      org.CreatedBy.Email,
		// TODO: Add org_url - requires app URL config which is not available in signal context
	})

	// TODO: Add analytics tracking for OrgCreated event
	// temporalanalytics.Writer is not available in the signal context.
	// Original code: w.analytics.Track(ctx, events.OrgCreated, map[string]any{"org_id": org.ID, "org_type": org.OrgType})

	// Add support users to trial orgs
	if hasTag(org.Tags, "Trial") {
		if _, err := activities.AwaitAddSupportUsersByOrgID(ctx, s.OrgID); err != nil {
			l := workflow.GetLogger(ctx)
			l.Error("unable to add support users to trial org",
				zap.Error(err),
				zap.String("org_id", s.OrgID))
		}
	}

	return nil
}

func (s *Signal) sendNotification(ctx workflow.Context, typ notifications.Type, orgID string, vars map[string]string) {
	l := workflow.GetLogger(ctx)

	if err := sharedactivities.AwaitSendEmail(ctx, sharedactivities.SendNotificationRequest{
		OrgID: orgID,
		Type:  typ,
		Vars:  vars,
	}); err != nil {
		l.Error("unable to send email",
			zap.Error(err),
			zap.String("type", typ.String()))
	}

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

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}
