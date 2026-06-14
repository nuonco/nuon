package githubevent

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "github_event"

type Signal struct {
	VCSConnectionEventID string `json:"vcs_connection_event_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.VCSConnectionEventID == "" {
		return errors.New("vcs_connection_event_id is required")
	}

	_, err := activities.AwaitGetVCSConnectionEvent(ctx, activities.GetVCSConnectionEventRequest{
		VCSConnectionEventID: s.VCSConnectionEventID,
	})
	if err != nil {
		return errors.Wrap(err, "vcs connection event not found")
	}

	return nil
}
