package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAppBranchRunRequest struct {
	AppBranchID            string
	AppBranchConfigID      string
	AppConfigID            string
	Force                  bool
	RunType                app.AppBranchRunType
	PlanOnly               bool
	EventType              string
	PRNumber               *int
	HeadSHA                string
	BaseBranch             string
	Labels                 labels.Labels
	TriggerEventDispatchID *string
	Preview                *PreviewRunInput
}

func (h *Helpers) CreateAppBranchRun(ctx context.Context, req *CreateAppBranchRunRequest) (*app.AppBranchRun, error) {
	runType := req.RunType
	if runType == "" {
		runType = app.AppBranchRunTypeManual
	}

	run := &app.AppBranchRun{
		AppBranchID:            req.AppBranchID,
		AppBranchConfigID:      req.AppBranchConfigID,
		TriggerEventDispatchID: req.TriggerEventDispatchID,
		AppConfigID:            req.AppConfigID,
		RunType:                runType,
		Force:                  req.Force,
		PlanOnly:               req.PlanOnly,
		EventType:              req.EventType,
		PRNumber:               req.PRNumber,
		HeadSHA:                req.HeadSHA,
		BaseBranch:             req.BaseBranch,
		Status:                 "pending",
		WorkflowID:             nil,
		Labeled:                labels.Labeled{Labels: req.Labels},
	}

	previewInput := MapLegacyPlanOnlyToPreviewInput(req)
	if previewInput != nil {
		var branchConfig app.AppBranchConfig
		configErr := h.db.WithContext(ctx).First(&branchConfig, "id = ?", req.AppBranchConfigID).Error
		var branch app.AppBranch
		branchErr := h.db.WithContext(ctx).First(&branch, "id = ?", req.AppBranchID).Error

		if req.Preview != nil {
			if configErr != nil {
				return nil, fmt.Errorf("unable to load app branch config for preview: %w", configErr)
			}
			if branchErr != nil {
				return nil, fmt.Errorf("unable to load app branch for preview: %w", branchErr)
			}
			preview, err := h.BuildAppBranchRunPreview(ctx, branch.AppID, &branchConfig, previewInput)
			if err != nil {
				return nil, err
			}
			run.Preview = preview
		} else if configErr == nil && branchErr == nil {
			preview, err := h.BuildAppBranchRunPreview(ctx, branch.AppID, &branchConfig, previewInput)
			if err == nil {
				run.Preview = preview
			}
		}
	}

	if run.Preview != nil {
		run.PlanOnly = true
	}

	if err := h.db.WithContext(ctx).Omit("Preview").Create(run).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", err)
	}

	if run.Preview != nil {
		run.Preview.AppBranchRunID = run.ID
		if err := h.db.WithContext(ctx).Create(run.Preview).Error; err != nil {
			return nil, fmt.Errorf("unable to create app branch run preview: %w", err)
		}
	}

	if shouldCreateComparison(runType, req.PlanOnly) || run.Preview != nil {
		if err := h.createAppBranchRunComparison(ctx, run); err != nil {
			return nil, fmt.Errorf("unable to create app branch run comparison: %w", err)
		}
	}

	return run, nil
}
