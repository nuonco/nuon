package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type CreateStepApprovalRequest struct {
	OwnerID   string `validate:"required"`
	OwnerType string `validate:"required"`

	RunnerJobID string
	StepID      string                       `validate:"required"`
	Type        app.WorkflowStepApprovalType `validate:"required"`

	Plan string
}

// @temporal-gen-v2 activity
func (a *Activities) CreateStepApproval(ctx context.Context, req *CreateStepApprovalRequest) (*app.WorkflowStepApproval, error) {
	plan := req.Plan
	if req.Plan == "" {
		job, err := a.GetJob(ctx, &GetJobRequest{
			ID: req.RunnerJobID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get job")
		}

		if job.Execution != nil && job.Execution.Result != nil {
			plan, err = job.Execution.Result.GetContentsDisplayString()
			if err != nil {
				return nil, errors.Wrap(err, "unable to get content display")
			}
		}

	}

	sa := app.WorkflowStepApproval{
		InstallWorkflowStepID: req.StepID,
		OwnerType:             req.OwnerType,
		OwnerID:               req.OwnerID,
		Contents:              plan,
		Type:                  req.Type,
	}

	if plan != "" {
		// the blob upload in WorkflowStepApproval's BeforeCreate hook requires org_id on the context
		if keys.OrgIDFromContext(ctx) == "" {
			var step app.WorkflowStep
			if res := a.db.WithContext(ctx).Select("org_id").First(&step, "id = ?", req.StepID); res.Error != nil {
				return nil, errors.Wrap(res.Error, "unable to look up step org for approval contents")
			}
			ctx = cctx.SetOrgIDContext(ctx, step.OrgID)
		}

		sa.ContentsBlob = &blobstore.Blob{}
		sa.ContentsBlob.Set(plan)
	}

	// workflows polymorphic step approvals do not have a runner job ID
	if req.RunnerJobID != "" {
		sa.RunnerJobID = generics.ToPtr(req.RunnerJobID)
	}

	// Soft-delete any existing approval for this step so the unique index
	// (install_workflow_step_id, deleted_at) allows the new row.
	if res := a.db.WithContext(ctx).
		Where("install_workflow_step_id = ? AND deleted_at = 0", req.StepID).
		Delete(&app.WorkflowStepApproval{}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to delete existing step approval")
	}

	res := a.db.WithContext(ctx).Create(&sa)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create step approval")
	}

	return &sa, nil
}
