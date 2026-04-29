package helpers

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// generateStepsSignal is a minimal signal type that matches the
// generateworkflowsteps.Signal type string. We define it here to avoid
// an import cycle (apps/helpers cannot import generateworkflowsteps).
type generateStepsSignal struct{}

func (s *generateStepsSignal) Type() qsignal.SignalType          { return "generate-workflow-steps" }
func (s *generateStepsSignal) Validate(_ workflow.Context) error { return nil }
func (s *generateStepsSignal) Execute(_ workflow.Context) error  { return nil }

func (s *Helpers) CreateWorkflow(ctx context.Context,
	appBranchID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
) (*app.Workflow, error) {
	metadata["app_branch_id"] = appBranchID
	return s.createWorkflow(ctx, appBranchID, "app_branches", workflowType, metadata, planOnly)
}

func (s *Helpers) CreateAppWorkflow(ctx context.Context,
	appID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
) (*app.Workflow, error) {
	return s.createWorkflow(ctx, appID, "apps", workflowType, metadata, planOnly)
}

func (s *Helpers) createWorkflow(ctx context.Context,
	ownerID, ownerType string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
) (*app.Workflow, error) {
	wf := app.Workflow{
		Type:              workflowType,
		OwnerID:           ownerID,
		OwnerType:         ownerType,
		Metadata:          generics.ToHstore(metadata),
		Status:            app.NewCompositeStatus(ctx, app.StatusPending),
		StepErrorBehavior: app.StepErrorBehaviorAbort,
		ApprovalOption:    app.InstallApprovalOptionPrompt,
		PlanOnly:          planOnly,
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: &generateStepsSignal{},
		},
	}

	res := s.db.WithContext(ctx).Create(&wf)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create workflow")
	}

	return &wf, nil
}
