// Package start defines the notebook-start queue signal. The notebook's queue
// dispatches it (at notebook-create time, and for recovery) to bring the warm
// per-notebook Temporal workflow online. Cell runs still dispatch to that
// workflow directly via update-with-start; this signal only owns its lifecycle.
package start

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "notebook_start"

type Signal struct {
	NotebookID string `json:"notebook_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.NotebookID == "" {
		return errors.New("notebook_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := activities.AwaitStartNotebookWorkflow(ctx, &activities.StartNotebookWorkflowRequest{
		NotebookID: s.NotebookID,
	}); err != nil {
		return errors.Wrap(err, "unable to start notebook workflow")
	}
	return nil
}
