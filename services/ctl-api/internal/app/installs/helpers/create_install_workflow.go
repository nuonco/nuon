package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (s *Helpers) CreateWorkflow(ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
) (*app.Workflow, error) {
	approvalOption := app.InstallApprovalOptionPrompt
	installConfig, err := s.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set approval option")
	}

	if installConfig != nil {
		approvalOption = installConfig.ApprovalOption
	}

	metadata["install_id"] = installID
	installWorkflow := app.Workflow{
		Type:      workflowType,
		OwnerID:   installID,
		OwnerType: "installs",
		Metadata:  generics.ToHstore(metadata),
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
		// DEPRECATED: for now we always abort on step errors
		StepErrorBehavior: app.StepErrorBehaviorAbort,
		ApprovalOption:    approvalOption,
		PlanOnly:          planOnly,
	}

	res := s.db.WithContext(ctx).Create(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	return &installWorkflow, nil
}
