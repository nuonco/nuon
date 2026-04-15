package healthcheck

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "vcs_connection_healthcheck"

type Signal struct {
	signal.Hooks
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

	// Check VCS connection health via GitHub APIs
	healthResult, err := activities.AwaitCheckVCSConnectionHealth(ctx, activities.CheckVCSConnectionHealthRequest{
		VCSConnectionID: s.VCSConnectionID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to check vcs connection health")
	}

	l.Info(fmt.Sprintf("vcs connection health check completed: status=%s repo_count=%d",
		healthResult.Status, healthResult.RepoCount))

	// Persist the status
	if err := activities.AwaitUpdateVCSConnectionStatus(ctx, activities.UpdateVCSConnectionStatusRequest{
		VCSConnectionID: s.VCSConnectionID,
		Status:          healthResult.Status,
		Description:     healthResult.Description,
		Metadata:        healthResult.Metadata,
	}); err != nil {
		return errors.Wrap(err, "unable to update vcs connection status")
	}

	return nil
}
