package bulkcancelworkflows

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/bulkcancel"
	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "general-bulk-cancel-workflows"

var _ signal.Signal = (*Signal)(nil)

type Signal struct {
	WorkflowIDs []string `json:"workflow_ids"`
	Reason      string   `json:"reason"`
}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	if len(s.WorkflowIDs) == 0 {
		return fmt.Errorf("workflow_ids is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, _ := log.WorkflowLogger(ctx)

	// unpinned, the child inherits this queue's task queue, where it is not registered
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: pkgworkflows.APITaskQueue,
	})

	if err := workflow.ExecuteChildWorkflow(childCtx, bulkcancel.WorkflowName, bulkcancel.Request{
		Pending: s.WorkflowIDs,
		Reason:  s.Reason,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("unable to bulk cancel workflows: %w", err)
	}

	if l != nil {
		l.Info("bulk workflow cancellation finished",
			zap.Int("requested", len(s.WorkflowIDs)),
			zap.String("reason", s.Reason))
	}

	return nil
}
