package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateWorkflowRequest struct {
	InstallID    string                `validate:"required"`
	WorkflowType app.WorkflowType      `validate:"required"`
	Metadata     map[string]string     `validate:"required"`
	ErrBehavior  app.StepErrorBehavior `validate:"required"`
	PlanOnly     bool                  `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*app.Workflow, error) {
	workflow, err := a.helpers.CreateWorkflow(ctx, req.InstallID, req.WorkflowType, req.Metadata, req.ErrBehavior, req.PlanOnly)
	if err != nil {
		return nil, fmt.Errorf("unable to create workflow: %w", err)
	}

	return workflow, nil
}
