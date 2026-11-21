package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	GitRef                 string
	BaseBranch             string
	IsDraftMode            bool
	Metadata               app.AppBranchRunMetadata
	Labels                 labels.Labels
	TriggerEventDispatchID *string
	Preview                *PreviewRunInput
	Request                *app.WorkflowRequest
}

func (h *Helpers) CreateAppBranchRun(ctx context.Context, req *CreateAppBranchRunRequest) (*app.AppBranchRun, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var branch app.AppBranch
	if err := h.db.WithContext(ctx).Preload("App").
		Where(&app.AppBranch{ID: req.AppBranchID, OrgID: orgID}, "ID", "OrgID").First(&branch).Error; err != nil {
		return nil, fmt.Errorf("unable to load app branch: %w", err)
	}
	if branch.App.ID == "" || branch.App.OrgID != orgID {
		return nil, fmt.Errorf("app branch does not belong to a live app in this org")
	}
	var branchConfig app.AppBranchConfig
	if err := h.db.WithContext(ctx).
		Where(&app.AppBranchConfig{ID: req.AppBranchConfigID, AppBranchID: branch.ID, OrgID: orgID}, "ID", "AppBranchID", "OrgID").First(&branchConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to load app branch config: %w", err)
	}
	previewInput := MapLegacyPlanOnlyToPreviewInput(req)
	configIDs := []string{req.AppConfigID}
	if previewInput != nil && previewInput.InputAppConfigID != req.AppConfigID {
		configIDs = append(configIDs, previewInput.InputAppConfigID)
	}
	for _, configID := range configIDs {
		if configID == "" {
			continue
		}
		var config app.AppConfig
		if err := h.db.WithContext(ctx).
			Where(&app.AppConfig{ID: configID, AppID: branch.AppID, OrgID: orgID}, "ID", "AppID", "OrgID").First(&config).Error; err != nil {
			return nil, fmt.Errorf("unable to load app config for branch run: %w", err)
		}
	}
	runType := req.RunType
	if runType == "" {
		runType = app.AppBranchRunTypeManual
	}

	metadata := req.Metadata
	if metadata.Trigger == "" {
		metadata.Trigger = triggerFromLegacyFields(req.EventType, runType)
	}
	if metadata.HeadSHA == "" {
		metadata.HeadSHA = req.HeadSHA
	}
	if metadata.GitRef == "" {
		metadata.GitRef = req.GitRef
	}
	if metadata.BaseBranch == "" {
		metadata.BaseBranch = req.BaseBranch
	}
	if metadata.PRNumber == nil {
		metadata.PRNumber = req.PRNumber
	}
	metadata.IsDraft = metadata.IsDraft || req.IsDraftMode

	run := &app.AppBranchRun{
		AppBranchID:            req.AppBranchID,
		AppBranchConfigID:      req.AppBranchConfigID,
		TriggerEventDispatchID: req.TriggerEventDispatchID,
		AppConfigID:            req.AppConfigID,
		RunType:                runType,
		Metadata:               metadata,
		Force:                  req.Force,
		PlanOnly:               req.PlanOnly,
		EventType:              req.EventType,
		PRNumber:               req.PRNumber,
		HeadSHA:                req.HeadSHA,
		BaseBranch:             req.BaseBranch,
		Status:                 "pending",
		WorkflowID:             nil,
		Labeled:                labels.Labeled{Labels: req.Labels},
		Request:                req.Request,
	}

	if previewInput != nil {
		preview, err := h.BuildAppBranchRunPreview(ctx, branch.AppID, &branchConfig, previewInput)
		if err != nil {
			if req.Preview != nil || runType == app.AppBranchRunTypeGitPreview {
				return nil, err
			}
		} else {
			run.Preview = preview
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

func triggerFromLegacyFields(eventType string, runType app.AppBranchRunType) app.AppBranchRunTrigger {
	switch eventType {
	case string(app.AppBranchRunTriggerPush):
		return app.AppBranchRunTriggerPush
	case string(app.AppBranchRunTriggerPullRequest):
		return app.AppBranchRunTriggerPullRequest
	case string(app.AppBranchRunTriggerTag):
		return app.AppBranchRunTriggerTag
	case string(app.AppBranchRunTriggerGithubLabel):
		return app.AppBranchRunTriggerGithubLabel
	case string(app.AppBranchRunTriggerOnboarding):
		return app.AppBranchRunTriggerOnboarding
	case string(app.AppBranchRunTriggerManual):
		return app.AppBranchRunTriggerManual
	}
	if runType == app.AppBranchRunTypeGitPreview {
		return app.AppBranchRunTriggerPullRequest
	}
	if runType == app.AppBranchRunTypeGit {
		return app.AppBranchRunTriggerPush
	}
	return app.AppBranchRunTriggerManual
}
