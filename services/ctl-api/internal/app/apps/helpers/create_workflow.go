package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (s *Helpers) CreateWorkflow(ctx context.Context,
	appBranchID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	errBehavior app.StepErrorBehavior,
	planOnly bool,
) (*app.Workflow, error) {
	approvalOption := app.InstallApprovalOptionPrompt

	metadata["app_branch_id"] = appBranchID
	installWorkflow := app.Workflow{
		Type:              workflowType,
		OwnerID:           appBranchID,
		OwnerType:         "app_branches",
		Metadata:          generics.ToHstore(metadata),
		Status:            app.NewCompositeStatus(ctx, app.StatusPending),
		StepErrorBehavior: errBehavior,
		ApprovalOption:    approvalOption,
		PlanOnly:          planOnly,
	}

	res := s.db.WithContext(ctx).Create(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	return &installWorkflow, nil
}
