package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (s *Helpers) CreateInstallFlow(ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	errBehavior app.StepErrorBehavior,
	planOnly bool,
) (*app.Workflow, error) {
	approvalOption := app.InstallApprovalOptionPrompt
	installConfig := app.InstallConfig{}
	resp := s.db.WithContext(ctx).Where("install_id = ?", installID).First(&installConfig)
	if resp.Error != nil && resp.Error != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(resp.Error, "unable to find install config")
	}

	if resp.Error != gorm.ErrRecordNotFound {
		approvalOption = installConfig.ApprovalOption
	}

	metadata["install_id"] = installID
	installWorkflow := app.Workflow{
		Type:              workflowType,
		InstallID:         pkggenerics.NewNullString(installID),
		OwnerID:           installID,
		OwnerType:         "installs",
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
