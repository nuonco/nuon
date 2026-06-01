package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (s *Helpers) CreateWorkflowWithRole(
	ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	role string,
) (*app.Workflow, error) {
	return s.createWorkflow(
		ctx,
		installID,
		workflowType,
		metadata,
		planOnly,
		role,
	)
}

func (s *Helpers) CreateWorkflow(
	ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
) (*app.Workflow, error) {
	return s.createWorkflow(
		ctx,
		installID,
		workflowType,
		metadata,
		planOnly,
		"",
	)
}

func (s *Helpers) createWorkflow(ctx context.Context,
	installID string,
	workflowType app.WorkflowType,
	metadata map[string]string,
	planOnly bool,
	role string,
) (*app.Workflow, error) {
	approvalOption := app.InstallApprovalOptionPrompt
	installConfig, err := s.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set approval option")
	}

	if installConfig != nil {
		approvalOption = installConfig.ApprovalOption
	}

	// Label the approval source so the UI can show where auto-approve came from
	if approvalOption == app.InstallApprovalOptionApproveAll {
		metadata["approval_type"] = "install-config"
	}

	metadata["install_id"] = installID
	var install app.Install
	if res := s.db.WithContext(ctx).Where("id = ?", installID).First(&install); res.Error == nil && install.Name != "" {
		metadata[app.WorkflowMetadataKeyOwnerName] = install.Name
	}

	status := app.NewCompositeStatus(ctx, app.StatusPending)
	if approvalOption == app.InstallApprovalOptionApproveAll {
		status.Metadata["approval_type"] = "install-config"
	}

	installWorkflow := app.Workflow{
		Type:      workflowType,
		OwnerID:   installID,
		OwnerType: "installs",
		Metadata:  generics.ToHstore(metadata),
		Status:    status,
		// DEPRECATED: for now we always abort on step errors
		StepErrorBehavior: app.StepErrorBehaviorAbort,
		ApprovalOption:    approvalOption,
		PlanOnly:          planOnly,
		Role:              role,
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: &generateworkflowsteps.Signal{},
		},
	}

	res := s.db.WithContext(ctx).Create(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	return &installWorkflow, nil
}
