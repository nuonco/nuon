package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateActionParams struct {
	AppID  string
	OrgID  string
	Name   string
	Labels map[string]string
}

func (h *Helpers) CreateAction(ctx context.Context, params *CreateActionParams) (*app.ActionWorkflow, error) {
	newAW := app.ActionWorkflow{
		AppID:   params.AppID,
		OrgID:   params.OrgID,
		Name:    params.Name,
		Status:  app.ActionWorkflowStatusActive,
		Labeled: labels.Labeled{Labels: labels.Labels(params.Labels)},
	}

	res := h.db.WithContext(ctx).Create(&newAW)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow: %w", res.Error)
	}

	return &newAW, nil
}
