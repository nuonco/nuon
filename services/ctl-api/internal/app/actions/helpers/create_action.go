package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateActionParams struct {
	AppID string
	OrgID string
	Name  string
}

func (h *Helpers) CreateAction(ctx context.Context, params *CreateActionParams) (*app.ActionWorkflow, error) {
	newAW := app.ActionWorkflow{
		AppID:  params.AppID,
		OrgID:  params.OrgID,
		Name:   params.Name,
		Status: app.ActionWorkflowStatusActive,
	}

	res := h.db.WithContext(ctx).Create(&newAW)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow: %w", res.Error)
	}

	if err := h.EnsureInstallAction(ctx, params.AppID, nil); err != nil {
		return nil, fmt.Errorf("unable to ensure install actions: %w", err)
	}

	return &newAW, nil
}
