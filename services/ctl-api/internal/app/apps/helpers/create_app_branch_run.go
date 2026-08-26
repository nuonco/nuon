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
}

func (h *Helpers) CreateAppBranchRun(ctx context.Context, req *CreateAppBranchRunRequest) (*app.AppBranchRun, error) {
	// Look up the previous successful run on the same branch for build diffing
	var previousRunID *string
	var prevRun app.AppBranchRun
	err := h.db.WithContext(ctx).
		Where(app.AppBranchRun{
			AppBranchID: req.AppBranchID,
			Status:      "success",
		}).
		Order("created_at DESC").
		First(&prevRun).Error
	if err == nil {
		previousRunID = &prevRun.ID
	}

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
		PreviousRunID:          previousRunID,
		Status:                 "pending",
		WorkflowID:             nil,
		Labeled:                labels.Labeled{Labels: req.Labels},
	}

	res := h.db.WithContext(ctx).Create(run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", res.Error)
	}

	if shouldCreateComparison(runType, req.PlanOnly) {
		if err := h.createAppBranchRunComparison(ctx, run); err != nil {
			return nil, fmt.Errorf("unable to create app branch run comparison: %w", err)
		}
	}

	return run, nil
}
