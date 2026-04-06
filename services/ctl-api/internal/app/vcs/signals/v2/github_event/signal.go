package githubevent

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "github_event"

type Signal struct {
	GithubEventID string `json:"github_event_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.GithubEventID == "" {
		return errors.New("github_event_id is required")
	}

	// Validate the event exists.
	_, err := activities.AwaitGetGithubEvent(ctx, activities.GetGithubEventRequest{
		GithubEventID: s.GithubEventID,
	})
	if err != nil {
		return errors.Wrap(err, "github event not found")
	}

	return nil
}
