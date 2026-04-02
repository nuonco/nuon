package updated

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "updated"

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Mark state as stale (copied from worker/updated.go)
	if err := activities.AwaitMarkStateStale(ctx, &activities.MarkStateStaleRequest{
		InstallID:       s.InstallID,
		TriggeredByID:   s.InstallID,
		TriggeredByType: "installs",
	}); err != nil {
		if !generics.IsGormErrRecordNotFound(err) {
			return errors.Wrap(err, "unable to mark state as stale")
		}
	}
	return nil
}
