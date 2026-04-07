package githubevent

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "github_event"

type Signal struct {
	VCSConnectionID string `json:"vcs_connection_id"`
	WebhookEventID  string `json:"webhook_event_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.VCSConnectionID == "" {
		return errors.New("vcs_connection_id is required")
	}
	if s.WebhookEventID == "" {
		return errors.New("webhook_event_id is required")
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

	result, err := activities.AwaitProcessGithubEvent(ctx, activities.ProcessGithubEventRequest{
		VCSConnectionID: s.VCSConnectionID,
		WebhookEventID:  s.WebhookEventID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to process github event")
	}

	l.Info(fmt.Sprintf("github event processed: event_type=%s commits_created=%d",
		result.EventType, result.CommitsCreated))

	return nil
}
