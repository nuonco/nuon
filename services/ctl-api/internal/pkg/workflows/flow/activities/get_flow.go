package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetFlowRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgWorkflowsFlowGetFlow(ctx context.Context, req GetFlowRequest) (*app.Workflow, error) {
	wf := app.Workflow{
		ID: req.ID,
	}
	if res := a.db.WithContext(ctx).
		First(&wf, "id = ?", req.ID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	return &wf, nil
}
