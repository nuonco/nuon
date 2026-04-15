package webhooksubscription

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "vcs_webhook_subscription"

type Signal struct {
	VCSConnectionID string `json:"vcs_connection_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.VCSConnectionID == "" {
		return errors.New("vcs_connection_id is required")
	}

	_, err := activities.AwaitGetVCSConnection(ctx, activities.GetVCSConnectionRequest{
		VCSConnectionID: s.VCSConnectionID,
	})
	if err != nil {
		return errors.Wrap(err, "vcs connection not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	result, err := activities.AwaitCreateWebhookSubscription(ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: s.VCSConnectionID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create webhook subscription")
	}

	if result.AlreadyExisted {
		l.Info(fmt.Sprintf("webhook subscription already exists for vcs connection %s: %s",
			s.VCSConnectionID, result.SubscriptionID))
	} else {
		l.Info(fmt.Sprintf("webhook subscription created for vcs connection %s: %s webhook_url=%s",
			s.VCSConnectionID, result.SubscriptionID, result.WebhookURL))
	}

	return nil
}
