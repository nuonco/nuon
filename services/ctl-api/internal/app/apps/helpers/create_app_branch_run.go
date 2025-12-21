package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAppBranchRunRequest struct {
	AppBranchID       string
	AppBranchConfigID string
	Force             bool
}

func (h *Helpers) CreateAppBranchRun(ctx context.Context, req *CreateAppBranchRunRequest) (*app.AppBranchRun, error) {
	run := &app.AppBranchRun{
		AppBranchID:       req.AppBranchID,
		AppBranchConfigID: req.AppBranchConfigID,
		Force:             req.Force,
		Status:            "pending",
		WorkflowID:        nil, // Set later after workflow creation
	}

	res := h.db.WithContext(ctx).Create(run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", res.Error)
	}

	return run, nil
}
