package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAppBranchRunRequest struct {
	AppBranchID       string  `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string  `json:"app_branch_config_id" validate:"required"`
	WorkflowID        *string `json:"workflow_id,omitempty"`
	Force             bool    `json:"force"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateAppBranchRun(ctx context.Context, req *CreateAppBranchRunRequest) (*app.AppBranchRun, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	var branch app.AppBranch
	if err := a.db.WithContext(ctx).First(&branch, "id = ?", req.AppBranchID).Error; err != nil {
		return nil, errors.Wrap(err, "app branch not found")
	}

	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).First(&config, "id = ?", req.AppBranchConfigID).Error; err != nil {
		return nil, errors.Wrap(err, "app branch config not found")
	}

	if req.WorkflowID != nil {
		var workflow app.Workflow
		if err := a.db.WithContext(ctx).First(&workflow, "id = ?", *req.WorkflowID).Error; err != nil {
			return nil, errors.Wrap(err, "workflow not found")
		}
	}

	run := app.AppBranchRun{
		AppBranchID:       req.AppBranchID,
		AppBranchConfigID: req.AppBranchConfigID,
		WorkflowID:        req.WorkflowID,
		Force:             req.Force,
		Status:            "pending",
	}

	if err := a.db.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, errors.Wrap(err, "unable to create app branch run")
	}

	return &run, nil
}
