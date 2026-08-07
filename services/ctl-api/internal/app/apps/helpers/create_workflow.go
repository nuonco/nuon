package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

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
	return s.createWorkflowWithDB(ctx, s.db, ownerID, ownerType, workflowType, metadata, planOnly, app.InstallApprovalOptionPrompt, "")
}

func (s *Helpers) createWorkflowWithDB(ctx context.Context, db *gorm.DB,
	ownerID, ownerType string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	approvalOption app.InstallApprovalOption,
	role string,
) (*app.Workflow, error) {
	status := app.NewCompositeStatus(ctx, app.StatusPending)
	if approvalOption == app.InstallApprovalOptionApproveAll {
		metadata["approval_type"] = "install-config"
		status.Metadata["approval_type"] = "install-config"
	}
	wf := app.Workflow{
		Type:              workflowType,
		OwnerID:           ownerID,
		OwnerType:         ownerType,
		Metadata:          generics.ToHstore(metadata),
		Status:            status,
		StepErrorBehavior: app.StepErrorBehaviorAbort,
		ApprovalOption:    approvalOption,
		PlanOnly:          planOnly,
		Role:              role,
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: &generateStepsSignal{},
		},
	}

	res := db.WithContext(ctx).Create(&wf)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create workflow")
	}

	return &wf, nil
}

func (s *Helpers) CreateWorkflowWithDB(ctx context.Context, db *gorm.DB, ownerID, ownerType string, workflowType app.WorkflowType, metadata map[string]string, planOnly bool, approvalOption app.InstallApprovalOption, role string) (*app.Workflow, error) {
	return s.createWorkflowWithDB(ctx, db, ownerID, ownerType, workflowType, metadata, planOnly, approvalOption, role)
}
